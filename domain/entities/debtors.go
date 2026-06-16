package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DebtorModel struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WorkspaceID        primitive.ObjectID `bson:"workspace_id,omitempty" json:"workspace_id"`
	PhoneNumber        string             `bson:"phone_number,omitempty" json:"phone_number"`
	Name               string             `bson:"name,omitempty" json:"name"`
	TotalDebt          float64            `bson:"total_debt,omitempty" json:"total_debt"`
	Status             string             `bson:"status,omitempty" json:"status"` // active/paid/defaulted/negotiating/pending
	ContactAttempts    int                `bson:"contact_attempts,omitempty" json:"contact_attempts"`
	SuccessfulContacts int                `bson:"successful_contacts,omitempty" json:"successful_contacts"`
	LastContactAt      time.Time          `bson:"last_contact_at,omitempty" json:"last_contact_at"`
	LastResponse       string             `bson:"last_response,omitempty" json:"last_response"`
	NextFollowUp       time.Time          `bson:"next_follow_up,omitempty" json:"next_follow_up"`
	Notes              string             `bson:"notes,omitempty" json:"notes"`
	CreatedAt          time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}
