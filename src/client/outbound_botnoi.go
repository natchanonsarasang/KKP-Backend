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
	tenantID    string
	accessToken string
}

type IOutboundBotnoiClient interface {
	// MakeCall dials req.Destination with req.AgentID over the trunk via
	// POST /v1/provisioning/tenants/{tenant_id}/calls. The debtor's data rides
	// along in req.Metadata, so no mid-call fetch is needed.
	//
	// It returns the call id Botnoi assigns to this call (its `call_id`), which
	// is echoed back as `conversation_id` in the result webhook and is the key we
	// correlate the webhook to its call_record on.
	MakeCall(req entities.OutboundCallRequest) (string, error)
}

// outboundCallResponse is the subset of Botnoi's place-a-call response we read.
// The id is returned as `call_id` (older responses used `outbound_id`); either
// one comes back as `conversation_id` in the result webhook.
type outboundCallResponse struct {
	CallID     string `json:"call_id"`
	OutboundID string `json:"outbound_id"`
}

// callID returns the usable correlation id, preferring call_id.
func (r outboundCallResponse) callID() string {
	if r.CallID != "" {
		return r.CallID
	}
	return r.OutboundID
}

func NewOutboundBotnoiClient() IOutboundBotnoiClient {
	return &OutboundBotnoiClient{
		client:      resty.New(),
		baseURL:     os.Getenv("OUTBOUND_URL"),
		tenantID:    os.Getenv("OUTBOUND_TENANT_ID"),
		accessToken: os.Getenv("OUTBOUND_ACCESS_TOKEN"),
	}
}

func (c *OutboundBotnoiClient) MakeCall(req entities.OutboundCallRequest) (string, error) {
	// destination is the clearest field to key log lines on for a single call.
	tag := "[Outbound " + req.Destination + "]"

	if c.baseURL == "" {
		fiberlog.Errorf("%s misconfigured: OUTBOUND_URL is empty", tag)
		return "", errors.New("outbound call failed: OUTBOUND_URL is not set")
	}
	if c.tenantID == "" {
		fiberlog.Errorf("%s misconfigured: OUTBOUND_TENANT_ID is empty", tag)
		return "", errors.New("outbound call failed: OUTBOUND_TENANT_ID is not set")
	}
	if c.accessToken == "" {
		fiberlog.Warnf("%s OUTBOUND_ACCESS_TOKEN is empty — request will likely be rejected", tag)
	}

	url := c.callsURL()
	fiberlog.Infof("%s placing call → destination=%s agent=%s url=%s", tag, req.Destination, req.AgentID, url)

	// Log the exact body we send so we can confirm what was really transmitted.
	// Note: this includes debtor PII (phone + metadata) — keep log access restricted.
	if body, marshalErr := json.Marshal(req); marshalErr == nil {
		fiberlog.Infof("%s payload=%s", tag, body)
	} else {
		fiberlog.Warnf("%s could not marshal payload for logging: %v", tag, marshalErr)
	}

	start := time.Now()
	resp, err := c.client.R().
		SetHeader("Authorization", "Bearer "+c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	elapsed := time.Since(start)

	if err != nil {
		// Transport-level failure (DNS, timeout, connection refused, ...).
		fiberlog.Errorf("%s request failed after %s: %v", tag, elapsed, err)
		return "", err
	}

	if resp.IsError() {
		// HTTP-level failure (non-2xx). Log status + body so the upstream reason is visible.
		fiberlog.Errorf("%s HTTP %d after %s: %s", tag, resp.StatusCode(), elapsed, resp.String())
		return "", errors.New("error response from Botnoi API (HTTP " + resp.Status() + "): " + resp.String())
	}

	// Parse out the call id so the webhook can be correlated to this call by it
	// (Botnoi echoes it as conversation_id). A missing id is not fatal — we fall
	// back to phone-number matching in the webhook — but it is worth flagging.
	var parsed outboundCallResponse
	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		fiberlog.Warnf("%s could not parse call id from response: %v (body=%s)", tag, err, resp.String())
	}
	callID := parsed.callID()
	if callID == "" {
		fiberlog.Warnf("%s response carried no call_id — webhook will fall back to phone matching (body=%s)", tag, resp.String())
	}

	fiberlog.Infof("%s success: HTTP %d in %s call_id=%s", tag, resp.StatusCode(), elapsed, callID)
	return callID, nil
}

// callsURL builds the place-a-call endpoint by appending the provisioning path
// to OUTBOUND_URL (the API base, e.g. ".../api"):
//
//	{OUTBOUND_URL}/v1/provisioning/tenants/{tenant_id}/calls
func (c *OutboundBotnoiClient) callsURL() string {
	base := strings.TrimRight(c.baseURL, "/")
	return base + "/v1/provisioning/tenants/" + c.tenantID + "/calls"
}
