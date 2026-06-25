package entities

import (
	"time"
)

type CallAttemptModel struct {
	ID              string             `bson:"id,omitempty" json:"id"`
	UserID          string             `bson:"user_id,omitempty" json:"user_id"`
	CallListItemID  string             `bson:"call_list_item_id,omitempty" json:"call_list_item_id"`
	CallRecordID    string             `bson:"call_record_id,omitempty" json:"call_record_id"`
	WorkspaceID     string             `bson:"workspace_id,omitempty" json:"workspace_id"`
	AttemptNumber   int                `bson:"attempt_number,omitempty" json:"attempt_number"`
	Status          string             `bson:"status,omitempty" json:"status"` // calling/finished
	CallOutcome     string             `bson:"call_outcome,omitempty" json:"call_outcome"`
	PickedUp        *bool               `bson:"picked_up,omitempty" json:"picked_up"`
	AiCategory      string             `bson:"ai_category,omitempty" json:"ai_category"`
	ConversationLog string             `bson:"conversation_log,omitempty" json:"conversation_log"`
	AudioURL        string             `bson:"audio_url,omitempty" json:"audio_url"`
	CallDuration    int                `bson:"call_duration,omitempty" json:"call_duration"`
	ErrorReason     string             `bson:"error_reason,omitempty" json:"error_reason"`
	CreatedAt       time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}

type CallAttemptFilter struct {
	WorkspaceID    string `json:"workspace_id"`
	UserID         string `json:"user_id"`
	CallListItemID string `json:"call_list_item_id"`
	Status         string `json:"status"`
	Limit          int64  `json:"limit"`
}
