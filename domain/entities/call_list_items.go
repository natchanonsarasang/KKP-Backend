package entities

import (
	"time"
)

type CallListItemModel struct {
	ID           string             `bson:"id,omitempty" json:"id"`
	UserID       string             `bson:"user_id,omitempty" json:"user_id"`
	DebtorID     string             `bson:"debtor_id,omitempty" json:"debtor_id"`
	WorkspaceID  string             `bson:"workspace_id,omitempty" json:"workspace_id"`
	TemplateID   string             `bson:"template_id,omitempty" json:"template_id"`
	ScheduledAt  time.Time          `bson:"scheduled_at,omitempty" json:"scheduled_at"`
	CalledAt     time.Time          `bson:"called_at,omitempty" json:"called_at"`
	Status       string             `bson:"status,omitempty" json:"status"` // pending/completed/failed
	CallRecordID string             `bson:"call_record_id,omitempty" json:"call_record_id"`
	CallOutcome  string             `bson:"call_outcome,omitempty" json:"call_outcome"`
	PickedUp     bool               `bson:"picked_up,omitempty" json:"picked_up"`
	Notes        string             `bson:"notes,omitempty" json:"notes"`
	CreatedAt    time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}

type CallListItemFilter struct {
	WorkspaceID   string    `json:"workspace_id"`
	UserID        string    `json:"user_id"`
	CalledAtGte   time.Time `json:"called_at_gte"`
	StatusesIn    []string  `json:"statuses_in"`
	StatusesNotIn []string  `json:"statuses_not_in"`
}
