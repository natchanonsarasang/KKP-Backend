package entities

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallSessionDataModel(t *testing.T) {
	now := time.Now().UTC()
	startedAt := now.Add(time.Minute)
	completedAt := now.Add(10 * time.Minute)
	errMsg := "some error occurred"
	settings := CallSessionSettings{
		MaxRetries:         3,
		DelayBetweenCalls:  10,
		ConcurrentCalls:    2,
		BusinessHoursOnly:  true,
		BusinessHoursStart: "09:00",
		BusinessHoursEnd:   "17:00",
		BusinessDays:       []int{1, 2, 3, 4, 5},
		TestMode:           false,
		TimezoneOffset:     420,
		Interruptible:      true,
	}

	session := CallSessionDataModel{
		ID:             "session-uuid-123",
		UserID:         "usr-1",
		WorkspaceID:    "ws-1",
		Status:         "running",
		TotalCalls:     10,
		CompletedCalls: 5,
		FailedCalls:    2,
		ConfirmedCalls: 3,
		TokenUsed:      150,
		Settings:       settings,
		ErrorMessage:   &errMsg,
		StartedAt:      &startedAt,
		CompletedAt:    &completedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Test JSON Marshalling
	data, err := json.Marshal(session)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON Unmarshalling
	var unmarshaled CallSessionDataModel
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	// Validate fields
	assert.Equal(t, session.ID, unmarshaled.ID)
	assert.Equal(t, session.UserID, unmarshaled.UserID)
	assert.Equal(t, session.WorkspaceID, unmarshaled.WorkspaceID)
	assert.Equal(t, session.Status, unmarshaled.Status)
	assert.Equal(t, session.TotalCalls, unmarshaled.TotalCalls)
	assert.Equal(t, session.CompletedCalls, unmarshaled.CompletedCalls)
	assert.Equal(t, session.FailedCalls, unmarshaled.FailedCalls)
	assert.Equal(t, session.ConfirmedCalls, unmarshaled.ConfirmedCalls)
	assert.Equal(t, session.TokenUsed, unmarshaled.TokenUsed)

	assert.Equal(t, session.Settings.MaxRetries, unmarshaled.Settings.MaxRetries)
	assert.Equal(t, session.Settings.DelayBetweenCalls, unmarshaled.Settings.DelayBetweenCalls)
	assert.Equal(t, session.Settings.ConcurrentCalls, unmarshaled.Settings.ConcurrentCalls)
	assert.Equal(t, session.Settings.BusinessHoursOnly, unmarshaled.Settings.BusinessHoursOnly)
	assert.Equal(t, session.Settings.BusinessHoursStart, unmarshaled.Settings.BusinessHoursStart)
	assert.Equal(t, session.Settings.BusinessHoursEnd, unmarshaled.Settings.BusinessHoursEnd)
	assert.Equal(t, session.Settings.BusinessDays, unmarshaled.Settings.BusinessDays)
	assert.Equal(t, session.Settings.TestMode, unmarshaled.Settings.TestMode)
	assert.Equal(t, session.Settings.TimezoneOffset, unmarshaled.Settings.TimezoneOffset)
	assert.Equal(t, session.Settings.Interruptible, unmarshaled.Settings.Interruptible)

	assert.NotNil(t, unmarshaled.ErrorMessage)
	assert.Equal(t, *session.ErrorMessage, *unmarshaled.ErrorMessage)

	assert.NotNil(t, unmarshaled.StartedAt)
	assert.True(t, session.StartedAt.Equal(*unmarshaled.StartedAt))

	assert.NotNil(t, unmarshaled.CompletedAt)
	assert.True(t, session.CompletedAt.Equal(*unmarshaled.CompletedAt))

	assert.True(t, session.CreatedAt.Equal(unmarshaled.CreatedAt))
	assert.True(t, session.UpdatedAt.Equal(unmarshaled.UpdatedAt))
}
