package services

import (
	"errors"
	"go-fiber-template/src/client"
	"os"
	"strings"
)

type botnoiLogsService struct {
	client client.IBotnoiLogsClient
}

// IBotnoiLogsService exposes the Botnoi conversation logs of the configured
// agent (OUTBOUND_AGENT_NAME). The agent is fixed server-side so callers can
// only read that agent's files.
type IBotnoiLogsService interface {
	ListFiles(startDate, endDate string) ([]byte, error)
	ReadLog(filePath string) ([]byte, error)
	GetAudio(filePath string) (body []byte, contentType string, contentDisposition string, err error)
}

var (
	ErrBotnoiAgentNotConfigured = errors.New("OUTBOUND_AGENT_NAME is not set")
	ErrBotnoiInvalidFilePath    = errors.New("file_path is not allowed")
)

func NewBotnoiLogsService() IBotnoiLogsService {
	return &botnoiLogsService{client: client.NewBotnoiLogsClient()}
}

func (sv *botnoiLogsService) ListFiles(startDate, endDate string) ([]byte, error) {
	agentName := os.Getenv("OUTBOUND_AGENT_NAME")
	if agentName == "" {
		return nil, ErrBotnoiAgentNotConfigured
	}
	return sv.client.ListFiles(agentName, startDate, endDate)
}

func (sv *botnoiLogsService) ReadLog(filePath string) ([]byte, error) {
	if err := validateBotnoiFilePath(filePath); err != nil {
		return nil, err
	}
	return sv.client.ReadLog(filePath)
}

func (sv *botnoiLogsService) GetAudio(filePath string) ([]byte, string, string, error) {
	if err := validateBotnoiFilePath(filePath); err != nil {
		return nil, "", "", err
	}
	return sv.client.GetAudio(filePath)
}

// validateBotnoiFilePath only allows files stored under "<agent_name>/", the
// folder layout Botnoi uses for list_file results.
func validateBotnoiFilePath(filePath string) error {
	agentName := os.Getenv("OUTBOUND_AGENT_NAME")
	if agentName == "" {
		return ErrBotnoiAgentNotConfigured
	}
	if filePath == "" || strings.Contains(filePath, "..") || !strings.HasPrefix(filePath, agentName+"/") {
		return ErrBotnoiInvalidFilePath
	}
	return nil
}
