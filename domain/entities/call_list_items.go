package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallListItemModel struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DebtorID     primitive.ObjectID `bson:"debtor_id,omitempty" json:"debtor_id"`
	WorkspaceID  primitive.ObjectID `bson:"workspace_id,omitempty" json:"workspace_id"`
	TemplateID   primitive.ObjectID `bson:"template_id,omitempty" json:"template_id"`
	ScheduledAt  time.Time          `bson:"scheduled_at,omitempty" json:"scheduled_at"`
	CalledAt     time.Time          `bson:"called_at,omitempty" json:"called_at"`
	Status       string             `bson:"status,omitempty" json:"status"` // pending/completed/failed
	CallRecordID primitive.ObjectID `bson:"call_record_id,omitempty" json:"call_record_id"`
	CallOutcome  string             `bson:"call_outcome,omitempty" json:"call_outcome"`
	PickedUp     bool               `bson:"picked_up,omitempty" json:"picked_up"`
	Notes        string             `bson:"notes,omitempty" json:"notes"`
	CreatedAt    time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}
