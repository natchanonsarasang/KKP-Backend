package client

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	fiberlog "github.com/gofiber/fiber/v2/log"
)

// BotnoiLogsClient reads conversation logs and recordings from the Botnoi
// Voicebot API (the "Conversation Logs" group in its docs). Responses are
// returned raw because the API does not document their shape.
type BotnoiLogsClient struct {
	client      *resty.Client
	baseURL     string
	accessToken string
}

type IBotnoiLogsClient interface {
	ListFiles(agentName, startDate, endDate string) ([]byte, error)
	ReadLog(filePath string) ([]byte, error)
	GetAudio(filePath string) (body []byte, contentType string, contentDisposition string, err error)
}

// BotnoiUpstreamError is a non-2xx answer from the Botnoi API.
type BotnoiUpstreamError struct {
	StatusCode int
	Body       string
}

func (e *BotnoiUpstreamError) Error() string {
	return fmt.Sprintf("botnoi api returned HTTP %d: %s", e.StatusCode, e.Body)
}

func NewBotnoiLogsClient() IBotnoiLogsClient {
	baseURL := os.Getenv("BOTNOI_API_URL")
	if baseURL == "" {
		baseURL = "https://voicebot-stg.botnoigroup.com/api"
	}
	return &BotnoiLogsClient{
		client:      resty.New().SetTimeout(30 * time.Second),
		baseURL:     strings.TrimRight(baseURL, "/"),
		accessToken: os.Getenv("OUTBOUND_ACCESS_TOKEN"),
	}
}

// ListFiles calls GET /logs/list_file/ (the trailing slash is required).
func (c *BotnoiLogsClient) ListFiles(agentName, startDate, endDate string) ([]byte, error) {
	req := c.newRequest().SetQueryParam("agent_name", agentName)
	if startDate != "" {
		req.SetQueryParam("start_date", startDate)
	}
	if endDate != "" {
		req.SetQueryParam("end_date", endDate)
	}
	resp, err := c.do(req, "/logs/list_file/")
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

// ReadLog calls GET /logs/read_log/ (the trailing slash is required).
func (c *BotnoiLogsClient) ReadLog(filePath string) ([]byte, error) {
	resp, err := c.do(c.newRequest().SetQueryParam("file_path", filePath), "/logs/read_log/")
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

// GetAudio calls GET /logs/audio_url/ with the conversation's .txt log path,
// the same way the Botnoi console does: with `Accept: audio/*` it answers with
// the recording bytes themselves. (GET /logs/audio 404s for these paths.)
func (c *BotnoiLogsClient) GetAudio(filePath string) ([]byte, string, string, error) {
	req := c.newRequest().SetHeader("Accept", "audio/*").SetQueryParam("file_path", filePath)
	resp, err := c.do(req, "/logs/audio_url/")
	if err != nil {
		return nil, "", "", err
	}
	return resp.Body(), resp.Header().Get("Content-Type"), resp.Header().Get("Content-Disposition"), nil
}

func (c *BotnoiLogsClient) newRequest() *resty.Request {
	return c.client.R().SetHeader("Authorization", "Bearer "+c.accessToken)
}

func (c *BotnoiLogsClient) do(req *resty.Request, path string) (*resty.Response, error) {
	tag := "[BotnoiLogs " + path + "]"
	if c.accessToken == "" {
		fiberlog.Warnf("%s OUTBOUND_ACCESS_TOKEN is empty — request will likely be rejected", tag)
	}

	start := time.Now()
	resp, err := req.Get(c.baseURL + path)
	elapsed := time.Since(start)

	if err != nil {
		fiberlog.Errorf("%s request failed after %s: %v", tag, elapsed, err)
		return nil, err
	}
	if resp.IsError() {
		fiberlog.Errorf("%s HTTP %d after %s: %s", tag, resp.StatusCode(), elapsed, resp.String())
		return nil, &BotnoiUpstreamError{StatusCode: resp.StatusCode(), Body: resp.String()}
	}

	fiberlog.Infof("%s success: HTTP %d in %s (%d bytes)", tag, resp.StatusCode(), elapsed, len(resp.Body()))
	return resp, nil
}
