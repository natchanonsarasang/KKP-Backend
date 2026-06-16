package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DebtorModel struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PhoneNumber       string             `bson:"phone_number" json:"phone_number"`
	Name              string             `bson:"name" json:"name"`
	TotalDebt         float64            `bson:"total_debt" json:"total_debt"`
	Status            string             `bson:"status" json:"status"` // active/paid/defaulted/negotiating/pending
	ContactAttempts   int                `bson:"contact_attempts" json:"contact_attempts"`
	SuccessfulContacts int                `bson:"successful_contacts" json:"successful_contacts"`
	LastContactAt     time.Time          `bson:"last_contact_at" json:"last_contact_at"`
	LastResponse      string             `bson:"last_response" json:"last_response"`
	NextFollowUp      time.Time          `bson:"next_follow_up" json:"next_follow_up"`
	Notes             string             `bson:"notes" json:"notes"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}
