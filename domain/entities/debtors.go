package entities

import (
	"time"
)

type DebtorModel struct {
	ID                 string            `bson:"id,omitempty" json:"id"`
	PhoneNumber        string            `bson:"phone_number,omitempty" json:"phone_number"`
	Name               string            `bson:"name,omitempty" json:"name"`
	LastName           string            `bson:"last_name,omitempty" json:"last_name"`
	TotalDebt          float64           `bson:"total_debt,omitempty" json:"total_debt"`
	Status             string            `bson:"status,omitempty" json:"status"` // active/paid/defaulted/negotiating/pending
	ContactAttempts    int               `bson:"contact_attempts,omitempty" json:"contact_attempts"`
	SuccessfulContacts int               `bson:"successful_contacts,omitempty" json:"successful_contacts"`
	LastContactAt      *time.Time        `bson:"last_contact_at,omitempty" json:"last_contact_at"`
	LastResponse       string            `bson:"last_response,omitempty" json:"last_response"`
	NextFollowUp       *time.Time        `bson:"next_follow_up,omitempty" json:"next_follow_up"`
	Notes              string            `bson:"notes,omitempty" json:"notes"`
	CreatedAt          time.Time         `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt          time.Time         `bson:"updated_at,omitempty" json:"updated_at"`
	AutoCallEnabled    bool              `bson:"auto_call_enabled,omitempty" json:"auto_call_enabled"`
	DueDate            *time.Time        `bson:"due_date,omitempty" json:"due_date"`
	CallAnswered       *bool             `bson:"call_answer,omitempty" json:"call_answered"`
	CallOutcome        string            `bson:"call_outcome,omitempty" json:"call_outcome"`
	PickedUpCount      int               `bson:"picked_up_count,omitempty" json:"picked_up_count"`
	NotPickedUpCount   int               `bson:"not_picked_up_count,omitempty" json:"not_picked_up_count"`
	Variables          map[string]string `bson:"variables,omitempty" json:"variables"`
	UserID             string            `bson:"user_id,omitempty" json:"user_id"`
	WorkspaceID        string            `bson:"workspace_id,omitempty" json:"workspace_id"`
	IsBlocked          bool              `bson:"is_blocked,omitempty" json:"is_blocked"`
	DateCon            string            `bson:"date_con,omitempty" json:"date_con"`
}

type DebtorStatsUpdate struct {
	ContactAttempts    int        `bson:"contact_attempts"`
	SuccessfulContacts int        `bson:"successful_contacts"`
	PickedUpCount      int        `bson:"picked_up_count"`
	NotPickedUpCount   int        `bson:"not_picked_up_count"`
	LastContactAt      *time.Time `bson:"last_contact_at"`
	LastResponse       string     `bson:"last_response"`
	CallOutcome        string     `bson:"call_outcome"`
	CallAnswered       *bool      `bson:"call_answer"`
}
