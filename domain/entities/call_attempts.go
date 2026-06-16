package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallAttemptModel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CallListItemID primitive.ObjectID `bson:"call_list_item_id" json:"call_list_item_id"`
	CallRecordID   primitive.ObjectID `bson:"call_record_id" json:"call_record_id"`
	WorkspaceID    primitive.ObjectID `bson:"workspace_id" json:"workspace_id"`
	AttemptNumber  int                `bson:"attempt_number" json:"attempt_number"`
	Status         string             `bson:"status" json:"status"` // calling/finished
	CallOutcome    string             `bson:"call_outcome" json:"call_outcome"`
	PickedUp       bool               `bson:"picked_up" json:"picked_up"`
	AiCategory     string             `bson:"ai_category" json:"ai_category"`
	ConversationLog string             `bson:"conversation_log" json:"conversation_log"`
	AudioURL       string             `bson:"audio_url" json:"audio_url"`
	CallDuration   int                `bson:"call_duration" json:"call_duration"`
	ErrorReason    string             `bson:"error_reason" json:"error_reason"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}
