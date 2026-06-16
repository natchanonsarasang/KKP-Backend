package entities

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallRecordDataModel(t *testing.T) {
	now := time.Now().UTC()
	dueDate := now.Add(24 * time.Hour)

	record := CallRecordDataModel{
		ID:              "c623c91c-1f5f-4027-a068-bd4f8286a111",
		TemplateID:      "tpl-123",
		PhoneNumber:     "+1234567890",
		AppointmentDate: "2026-06-16",
		AppointmentTime: "11:15:00",
		Status:          StatusConfirmed,
		BotnoiCallID:    "botnoi-call-xyz",
		ResultData:      map[string]any{"key": "value"},
		DueDate:         dueDate,
		Amount:          150.50,
		UserID:          "usr-999",
		WorkspaceID:     "ws-777",
		CallDuration:    120,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Test JSON Marshalling
	data, err := json.Marshal(record)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON Unmarshalling
	var unmarshaled CallRecordDataModel
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	// Validate fields
	assert.Equal(t, record.ID, unmarshaled.ID)
	assert.Equal(t, record.TemplateID, unmarshaled.TemplateID)
	assert.Equal(t, record.PhoneNumber, unmarshaled.PhoneNumber)
	assert.Equal(t, record.AppointmentDate, unmarshaled.AppointmentDate)
	assert.Equal(t, record.AppointmentTime, unmarshaled.AppointmentTime)
	assert.Equal(t, record.Status, unmarshaled.Status)
	assert.Equal(t, record.BotnoiCallID, unmarshaled.BotnoiCallID)

	// Interface map assertion requires type assertion check or comparison
	resultDataMap, ok := unmarshaled.ResultData.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", resultDataMap["key"])

	assert.True(t, record.DueDate.Equal(unmarshaled.DueDate))
	assert.Equal(t, record.Amount, unmarshaled.Amount)
	assert.Equal(t, record.UserID, unmarshaled.UserID)
	assert.Equal(t, record.WorkspaceID, unmarshaled.WorkspaceID)
	assert.Equal(t, record.CallDuration, unmarshaled.CallDuration)
	assert.True(t, record.CreatedAt.Equal(unmarshaled.CreatedAt))
	assert.True(t, record.UpdatedAt.Equal(unmarshaled.UpdatedAt))
}

func TestCallStatusEnum(t *testing.T) {
	assert.Equal(t, CallStatus("confirmed"), StatusConfirmed)
	assert.Equal(t, CallStatus("declined"), StatusDeclined)
	assert.Equal(t, CallStatus("no_response"), StatusNoResponse)
	assert.Equal(t, CallStatus("no_answer"), StatusNoAnswer)
	assert.Equal(t, CallStatus("hanged_up"), StatusHangedUp)
	assert.Equal(t, CallStatus("pending"), StatusPending)
	assert.Equal(t, CallStatus("completed"), StatusCompleted)
	assert.Equal(t, CallStatus("busy"), StatusBusy)
	assert.Equal(t, CallStatus("failed"), StatusFailed)
	assert.Equal(t, CallStatus("rejected"), StatusRejected)
	assert.Equal(t, CallStatus("voicemail"), StatusVoicemail)
}
