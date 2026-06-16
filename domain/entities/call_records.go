package entities

import (
	"time"

	"github.com/google/uuid"
)

// CallStatus represents the status of a call record.
type CallStatus string

const (
	StatusConfirmed  CallStatus = "confirmed"
	StatusDeclined   CallStatus = "declined"
	StatusNoResponse CallStatus = "no_response"
	StatusNoAnswer   CallStatus = "no_answer"
	StatusHangedUp   CallStatus = "hanged_up"
	StatusPending    CallStatus = "pending"
	StatusCompleted  CallStatus = "completed"
	StatusBusy       CallStatus = "busy"
	StatusFailed     CallStatus = "failed"
	StatusRejected   CallStatus = "rejected"
	StatusVoicemail  CallStatus = "voicemail"
	StatusCalling    CallStatus = "calling"
)

// CallRecordDataModel represents a call record in the database.
type CallRecordDataModel struct {
	ID              string        `json:"id" bson:"id,omitempty"`
	TemplateID      *string       `json:"template_id" bson:"template_id,omitempty"`
	PhoneNumber     string        `json:"phone_number" bson:"phone_number,omitempty"`
	AppointmentDate string        `json:"appointment_date" bson:"appointment_date,omitempty"`
	AppointmentTime string        `json:"appointment_time" bson:"appointment_time,omitempty"`
	Status          CallStatus    `json:"status" bson:"status,omitempty"`
	BotnoiCallID    string        `json:"botnoi_call_id" bson:"botnoi_call_id,omitempty"`
	ResultData      *interface{}  `json:"result_data" bson:"result_data,omitempty"`
	DueDate         time.Time     `json:"due_date" bson:"due_date,omitempty"`
	Amount          float64       `json:"amount" bson:"amount,omitempty"`
	UserID          string        `json:"user_id" bson:"user_id,omitempty"`
	WorkspaceID     string        `json:"workspace_id" bson:"workspace_id,omitempty"`
	CallDuration    int           `json:"call_duration" bson:"call_duration,omitempty"`
	CreatedAt       time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt       time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
}

// NewCallRecord initializes a new CallRecordDataModel with a UUIDv4 ID and current timestamps.
func NewCallRecord() CallRecordDataModel {
	now := time.Now().UTC()
	return CallRecordDataModel{
		ID:        uuid.NewString(),
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type CallRecordFilter struct {
	UserID       string `json:"user_id"`
	WorkspaceID  string `json:"workspace_id"`
	Status       string `json:"status"`
	BotnoiCallID string `json:"botnoi_call_id"`
}
