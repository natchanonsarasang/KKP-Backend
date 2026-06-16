package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallAttemptModel struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CallListItemID  primitive.ObjectID `bson:"call_list_item_id,omitempty" json:"call_list_item_id"`
	CallRecordID    primitive.ObjectID `bson:"call_record_id,omitempty" json:"call_record_id"`
	WorkspaceID     primitive.ObjectID `bson:"workspace_id,omitempty" json:"workspace_id"`
	AttemptNumber   int                `bson:"attempt_number,omitempty" json:"attempt_number"`
	Status          string             `bson:"status,omitempty" json:"status"` // calling/finished
	CallOutcome     string             `bson:"call_outcome,omitempty" json:"call_outcome"`
	PickedUp        bool               `bson:"picked_up,omitempty" json:"picked_up"`
	AiCategory      string             `bson:"ai_category,omitempty" json:"ai_category"`
	ConversationLog string             `bson:"conversation_log,omitempty" json:"conversation_log"`
	AudioURL        string             `bson:"audio_url,omitempty" json:"audio_url"`
	CallDuration    int                `bson:"call_duration,omitempty" json:"call_duration"`
	ErrorReason     string             `bson:"error_reason,omitempty" json:"error_reason"`
	CreatedAt       time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}
