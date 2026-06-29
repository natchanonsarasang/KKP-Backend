package client

import (
	"errors"
	"go-fiber-template/domain/entities"
	"os"
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

	fiberlog.Infof("%s placing call → phone=%s url=%s", tag, payload.PhoneNumber, c.baseURL)

	start := time.Now()
	resp, err := c.client.R().
		SetHeader("Authorization", "Bearer "+c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(c.baseURL)
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
