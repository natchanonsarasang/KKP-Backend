package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"go-fiber-template/src/client"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type botnoiLogsService struct {
	client                  client.IBotnoiLogsClient
	CallRecordsRepository   repositories.ICallRecordsRepository
	CallListItemsRepository repositories.ICallListItemsRepository
}

// IBotnoiLogsService exposes the Botnoi conversation logs of the configured
// agent (OUTBOUND_AGENT_ID). The agent is fixed server-side so callers can
// only read that agent's files.
type IBotnoiLogsService interface {
	ListFiles(startDate, endDate string) ([]byte, error)
	ReadLog(filePath string) ([]byte, error)
	GetAudio(filePath string) (body []byte, contentType string, contentDisposition string, err error)
	// LookupConversationsByUser links Botnoi conversation ids to the user's own
	// call records and their debtors. Ids with no matching record are omitted.
	LookupConversationsByUser(userID string, conversationIDs []string) ([]entities.BotnoiConversationLink, error)
}

var (
	ErrBotnoiAgentIDNotConfigured = errors.New("OUTBOUND_AGENT_ID is not set")
	ErrBotnoiInvalidFilePath      = errors.New("file_path is not allowed")
)

func NewBotnoiLogsService(callRecordsRepo repositories.ICallRecordsRepository, callListItemsRepo repositories.ICallListItemsRepository) IBotnoiLogsService {
	return &botnoiLogsService{
		client:                  client.NewBotnoiLogsClient(),
		CallRecordsRepository:   callRecordsRepo,
		CallListItemsRepository: callListItemsRepo,
	}
}

func (sv *botnoiLogsService) LookupConversationsByUser(userID string, conversationIDs []string) ([]entities.BotnoiConversationLink, error) {
	records, err := sv.CallRecordsRepository.FindByConversationIDsByUser(userID, conversationIDs)
	if err != nil {
		return nil, err
	}

	recordIDs := make([]string, 0, len(*records))
	for _, r := range *records {
		recordIDs = append(recordIDs, r.ID)
	}
	items, err := sv.CallListItemsRepository.FindByCallRecordIDs(recordIDs)
	if err != nil {
		return nil, err
	}
	itemByRecord := map[string]entities.CallListItemModel{}
	for _, item := range *items {
		itemByRecord[item.CallRecordID] = item
	}

	links := make([]entities.BotnoiConversationLink, 0, len(*records))
	for _, r := range *records {
		link := entities.BotnoiConversationLink{
			ConversationID: conversationIDOf(r),
			CallRecordID:   r.ID,
			DebtorPhone:    r.PhoneNumber,
			Status:         string(r.Status),
			CallDuration:   r.CallDuration,
		}
		if item, ok := itemByRecord[r.ID]; ok {
			link.DebtorName = item.DebtorName
			if item.DebtorPhone != "" {
				link.DebtorPhone = item.DebtorPhone
			}
			link.CallOutcome = item.CallOutcome
			link.AICategory = item.AICategory
		}
		links = append(links, link)
	}
	return links, nil
}

// conversationIDOf reads the record's Botnoi conversation id, falling back to
// the webhook payload stored in result_data for records saved before the
// conversation_id field existed.
func conversationIDOf(r entities.CallRecordDataModel) string {
	if r.ConversationID != "" {
		return r.ConversationID
	}
	if r.ResultData == nil {
		return ""
	}
	if doc, ok := (*r.ResultData).(bson.D); ok {
		for _, e := range doc {
			if e.Key == "conversationid" {
				if s, ok := e.Value.(string); ok {
					return s
				}
			}
		}
	}
	return ""
}

func (sv *botnoiLogsService) ListFiles(startDate, endDate string) ([]byte, error) {
	// The logs API filters by agent_id, not agent_name (see BotnoiLogsClient.ListFiles).
	agentID := os.Getenv("OUTBOUND_AGENT_ID")
	if agentID == "" {
		return nil, ErrBotnoiAgentIDNotConfigured
	}
	return sv.client.ListFiles(agentID, startDate, endDate)
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

// validateBotnoiFilePath only allows files stored under "<agent_id>/", the
// folder layout Botnoi uses for list_file results (e.g. "agt_3d2fa3ed2358/...").
func validateBotnoiFilePath(filePath string) error {
	agentID := os.Getenv("OUTBOUND_AGENT_ID")
	if agentID == "" {
		return ErrBotnoiAgentIDNotConfigured
	}
	if filePath == "" || strings.Contains(filePath, "..") || !strings.HasPrefix(filePath, agentID+"/") {
		return ErrBotnoiInvalidFilePath
	}
	return nil
}
