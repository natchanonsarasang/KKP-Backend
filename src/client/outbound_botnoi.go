package client

import (
	"encoding/json"
	"errors"
	"go-fiber-template/domain/entities"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	fiberlog "github.com/gofiber/fiber/v2/log"
)

type OutboundBotnoiClient struct {
	client      *resty.Client
	baseURL     string
	accessToken string
}

type IOutboundBotnoiClient interface {
	MakeCall(payload entities.OutboundBotnoiDataModel) error
	// CreateBatch creates an outbound batch (POST /outbound/batch) and returns the
	// batch_id Botnoi assigns. That batch_id is then used as the outbound_id when
	// placing the actual call.
	CreateBatch(phoneNumber, batchName string) (string, error)
	// GetBatchStatus fetches a batch (GET /outbound/batch/{id}). While the webhook
	// is unavailable we poll this to learn a call's outcome from its Calls list.
	GetBatchStatus(batchID string) (*entities.OutboundBatchStatusResponse, error)
}

func NewOutboundBotnoiClient(token string, host string, port string) IOutboundBotnoiClient {
	return &OutboundBotnoiClient{
		client:      resty.New(),
		baseURL:     os.Getenv("OUTBOUND_URL"),
		accessToken: os.Getenv("OUTBOUND_ACCESS_TOKEN"),
	}
}

func (c *OutboundBotnoiClient) MakeCall(payload entities.OutboundBotnoiDataModel) error {
	// outbound_id is the correlation key echoed back by the webhook, so it is the
	// most useful field to key log lines on when tracing a single call.
	tag := "[Outbound " + payload.OutboundID + "]"

	if c.baseURL == "" {
		fiberlog.Errorf("%s misconfigured: OUTBOUND_URL is empty", tag)
		return errors.New("outbound call failed: OUTBOUND_URL is not set")
	}
	if c.accessToken == "" {
		fiberlog.Warnf("%s OUTBOUND_ACCESS_TOKEN is empty — request will likely be rejected", tag)
	}

	url := c.outboundURL("")
	fiberlog.Infof("%s placing call → phone=%s agent=%s url=%s", tag, payload.TelephoneNumber, payload.AgentName, url)

	// Log the exact body we send so we can confirm what was really transmitted.
	// Note: this includes debtor PII (phone, TTS variables) — keep log access restricted.
	if body, marshalErr := json.Marshal(payload); marshalErr == nil {
		fiberlog.Infof("%s payload=%s", tag, body)
	} else {
		fiberlog.Warnf("%s could not marshal payload for logging: %v", tag, marshalErr)
	}

	start := time.Now()
	resp, err := c.client.R().
		SetHeader("Authorization", "Bearer "+c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(url)
	elapsed := time.Since(start)

	if err != nil {
		// Transport-level failure (DNS, timeout, connection refused, ...).
		fiberlog.Errorf("%s request failed after %s: %v", tag, elapsed, err)
		return err
	}

	if resp.IsError() {
		// HTTP-level failure (non-2xx). Log status + body so the upstream reason is visible.
		fiberlog.Errorf("%s HTTP %d after %s: %s", tag, resp.StatusCode(), elapsed, resp.String())
		return errors.New("error response from Botnoi API (HTTP " + resp.Status() + "): " + resp.String())
	}

	fiberlog.Infof("%s success: HTTP %d in %s", tag, resp.StatusCode(), elapsed)
	return nil
}

// outboundURL builds an outbound endpoint from the configured OUTBOUND_URL, which
// is the API base (e.g. ".../api"). The "/outbound" segment is appended here:
//   outboundURL("")            → ".../api/outbound"        (place call)
//   outboundURL("batch")       → ".../api/outbound/batch"  (create batch)
//   outboundURL("batch/<id>")  → ".../api/outbound/batch/<id>" (poll batch)
func (c *OutboundBotnoiClient) outboundURL(suffix string) string {
	base := strings.TrimRight(c.baseURL, "/") + "/outbound"
	if suffix != "" {
		base += "/" + suffix
	}
	return base
}

func (c *OutboundBotnoiClient) CreateBatch(phoneNumber, batchName string) (string, error) {
	tag := "[Outbound batch " + batchName + "]"

	if c.baseURL == "" {
		fiberlog.Errorf("%s misconfigured: OUTBOUND_URL is empty", tag)
		return "", errors.New("create batch failed: OUTBOUND_URL is not set")
	}

	url := c.outboundURL("batch")
	reqBody := entities.OutboundBatchCreateRequest{
		PhoneNumbers: []string{phoneNumber},
		AgentName:    os.Getenv("OUTBOUND_AGENT_NAME"),
		BatchName:    batchName,
	}

	if body, marshalErr := json.Marshal(reqBody); marshalErr == nil {
		fiberlog.Infof("%s creating batch url=%s payload=%s", tag, url, body)
	}

	var result entities.OutboundBatchCreateResponse
	start := time.Now()
	resp, err := c.client.R().
		SetHeader("Authorization", "Bearer "+c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		SetResult(&result).
		Post(url)
	elapsed := time.Since(start)

	if err != nil {
		fiberlog.Errorf("%s request failed after %s: %v", tag, elapsed, err)
		return "", err
	}
	if resp.IsError() {
		fiberlog.Errorf("%s HTTP %d after %s: %s", tag, resp.StatusCode(), elapsed, resp.String())
		return "", errors.New("error response from Botnoi batch API (HTTP " + resp.Status() + "): " + resp.String())
	}
	if result.BatchID == "" {
		fiberlog.Errorf("%s response had no batch_id: %s", tag, resp.String())
		return "", errors.New("create batch failed: response contained no batch_id")
	}

	fiberlog.Infof("%s created: batch_id=%s in %s", tag, result.BatchID, elapsed)
	return result.BatchID, nil
}

func (c *OutboundBotnoiClient) GetBatchStatus(batchID string) (*entities.OutboundBatchStatusResponse, error) {
	tag := "[Outbound batch " + batchID + "]"

	if c.baseURL == "" {
		return nil, errors.New("get batch failed: OUTBOUND_URL is not set")
	}

	url := c.outboundURL("batch/" + batchID)
	var result entities.OutboundBatchStatusResponse
	resp, err := c.client.R().
		SetHeader("Authorization", "Bearer "+c.accessToken).
		SetResult(&result).
		Get(url)

	if err != nil {
		fiberlog.Errorf("%s status request failed: %v", tag, err)
		return nil, err
	}
	if resp.IsError() {
		fiberlog.Errorf("%s status HTTP %d: %s", tag, resp.StatusCode(), resp.String())
		return nil, errors.New("error response from Botnoi batch API (HTTP " + resp.Status() + "): " + resp.String())
	}

	return &result, nil
}
