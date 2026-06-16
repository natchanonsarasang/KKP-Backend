package entities

import (
	"time"

	"github.com/google/uuid"
)

// CallSessionDataModel represents a call campaign/session run in the database.
type CallSessionDataModel struct {
	ID             string        `json:"id" bson:"id,omitempty"`
	UserID         string        `json:"user_id" bson:"user_id,omitempty"`
	WorkspaceID    string        `json:"workspace_id" bson:"workspace_id,omitempty"`
	Status         string        `json:"status" bson:"status,omitempty"`
	TotalCalls     int           `json:"total_calls" bson:"total_calls,omitempty"`
	CompletedCalls int           `json:"completed_calls" bson:"completed_calls,omitempty"`
	FailedCalls    int           `json:"failed_calls" bson:"failed_calls,omitempty"`
	ConfirmedCalls int           `json:"confirmed_calls" bson:"confirmed_calls,omitempty"`
	TokenUsed      int           `json:"token_used" bson:"token_used,omitempty"`
	Settings       *interface{}  `json:"settings" bson:"settings,omitempty"`
	ErrorMessage   *string       `json:"error_message" bson:"error_message,omitempty"`
	StartedAt      *time.Time    `json:"started_at" bson:"started_at,omitempty"`
	CompletedAt    *time.Time    `json:"completed_at" bson:"completed_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt      time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
}

// NewCallSession initializes a new CallSessionDataModel with a UUIDv4 ID and current timestamps.
func NewCallSession() CallSessionDataModel {
	now := time.Now().UTC()
	return CallSessionDataModel{
		ID:        uuid.NewString(),
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type CallSessionFilter struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
}
