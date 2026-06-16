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
	var settings interface{} = map[string]interface{}{"retry_count": float64(3)}

	session := CallSessionDataModel{
		ID:             "session-uuid-123",
		UserID:          "usr-1",
		WorkspaceID:     "ws-1",
		Status:         "running",
		TotalCalls:     10,
		CompletedCalls: 5,
		FailedCalls:    2,
		ConfirmedCalls: 3,
		TokenUsed:      150,
		Settings:       &settings,
		ErrorMessage:   &errMsg,
		StartedAt:      &startedAt,
		CompletedAt:    &completedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
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

	assert.NotNil(t, unmarshaled.Settings)
	settingsMap, ok := (*unmarshaled.Settings).(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(3), settingsMap["retry_count"])

	assert.NotNil(t, unmarshaled.ErrorMessage)
	assert.Equal(t, *session.ErrorMessage, *unmarshaled.ErrorMessage)

	assert.NotNil(t, unmarshaled.StartedAt)
	assert.True(t, session.StartedAt.Equal(*unmarshaled.StartedAt))

	assert.NotNil(t, unmarshaled.CompletedAt)
	assert.True(t, session.CompletedAt.Equal(*unmarshaled.CompletedAt))

	assert.True(t, session.CreatedAt.Equal(unmarshaled.CreatedAt))
	assert.True(t, session.UpdatedAt.Equal(unmarshaled.UpdatedAt))
}
