package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallListItemModel struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DebtorID     primitive.ObjectID `bson:"debtor_id" json:"debtor_id"`
	WorkspaceID  primitive.ObjectID `bson:"workspace_id" json:"workspace_id"`
	TemplateID   primitive.ObjectID `bson:"template_id" json:"template_id"`
	ScheduledAt  time.Time          `bson:"scheduled_at" json:"scheduled_at"`
	CalledAt     time.Time          `bson:"called_at" json:"called_at"`
	Status       string             `bson:"status" json:"status"` // pending/completed/failed
	CallRecordID primitive.ObjectID `bson:"call_record_id" json:"call_record_id"`
	CallOutcome  string             `bson:"call_outcome" json:"call_outcome"`
	PickedUp     bool               `bson:"picked_up" json:"picked_up"`
	Notes        string             `bson:"notes" json:"notes"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
