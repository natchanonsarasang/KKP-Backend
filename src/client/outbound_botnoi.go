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
	MakeCall(req entities.OutboundCallRequest) error
}

func NewOutboundBotnoiClient() IOutboundBotnoiClient {
	return &OutboundBotnoiClient{
		client:      resty.New(),
		baseURL:     os.Getenv("OUTBOUND_URL"),
		tenantID:    os.Getenv("OUTBOUND_TENANT_ID"),
		accessToken: os.Getenv("OUTBOUND_ACCESS_TOKEN"),
	}
}

func (c *OutboundBotnoiClient) MakeCall(req entities.OutboundCallRequest) error {
	// destination is the clearest field to key log lines on for a single call.
	tag := "[Outbound " + req.Destination + "]"

	if c.baseURL == "" {
		fiberlog.Errorf("%s misconfigured: OUTBOUND_URL is empty", tag)
		return errors.New("outbound call failed: OUTBOUND_URL is not set")
	}
	if c.tenantID == "" {
		fiberlog.Errorf("%s misconfigured: OUTBOUND_TENANT_ID is empty", tag)
		return errors.New("outbound call failed: OUTBOUND_TENANT_ID is not set")
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

// callsURL builds the place-a-call endpoint from OUTBOUND_URL. OUTBOUND_URL is
// the Botnoi host; a legacy trailing "/api" segment is stripped since the
// provisioning endpoint lives at the host root:
//
//	{host}/v1/provisioning/tenants/{tenant_id}/calls
func (c *OutboundBotnoiClient) callsURL() string {
	base := strings.TrimRight(c.baseURL, "/")
	base = strings.TrimSuffix(base, "/api")
	return base + "/v1/provisioning/tenants/" + c.tenantID + "/calls"
}
