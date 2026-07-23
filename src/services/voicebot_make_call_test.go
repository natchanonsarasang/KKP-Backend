package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockOutboundBotnoiClient struct {
	MakeCallFunc func(payload entities.OutboundBotnoiDataModel) error
}

func (m *mockOutboundBotnoiClient) MakeCall(payload entities.OutboundBotnoiDataModel) error {
	if m.MakeCallFunc != nil {
		return m.MakeCallFunc(payload)
	}
	return nil
}

func TestVoicebotMakeCallService_Validation(t *testing.T) {
	mockClient := &mockOutboundBotnoiClient{}
	svc := &voicebotMakeCallService{
		outboutClient: mockClient,
	}

	// Case 1: Empty phone number
	err := svc.MakeCall(entities.VoicebotMakeCallDataModel{
		PhoneNumber: "",
		Variables:   map[string]any{"name": "John"},
	})
	assert.Error(t, err)
	assert.Equal(t, "phone_number is required", err.Error())

	// Case 2: Nil variables map — no longer an error; the call proceeds with
	// the mock defaults from defaultCallVariables.
	var capturedPayload entities.OutboundBotnoiDataModel
	mockClient.MakeCallFunc = func(payload entities.OutboundBotnoiDataModel) error {
		capturedPayload = payload
		return nil
	}
	err = svc.MakeCall(entities.VoicebotMakeCallDataModel{
		PhoneNumber: "0909722021",
		Variables:   nil,
	})
	assert.NoError(t, err)
	assert.Contains(t, capturedPayload.Flow, "<!customer_name|คุณสมชาย!>")
	assert.Contains(t, capturedPayload.Flow, "<!car_detail|กก1111!>")
	assert.Contains(t, capturedPayload.Flow, "<!province|กรุงเทพมหานคร!>")
	assert.Contains(t, capturedPayload.Flow, "<!total_debt|3000!>")
	assert.Contains(t, capturedPayload.Flow, "<!overdue_installment|2!>")
}

func TestVoicebotMakeCallService_MakeCall(t *testing.T) {
	var capturedPayload entities.OutboundBotnoiDataModel
	mockClient := &mockOutboundBotnoiClient{
		MakeCallFunc: func(payload entities.OutboundBotnoiDataModel) error {
			capturedPayload = payload
			return nil
		},
	}

	svc := &voicebotMakeCallService{
		outboutClient: mockClient,
	}

	variables := map[string]any{
		"name":                "สมชาย",
		"car_detail":          "Toyota Vios กข1234",
		"total_debt":          1500.50,
		"total_interest":      120.25,
		"total_fine":          50,
		"overdue_installment": "3",
	}

	err := svc.MakeCall(entities.VoicebotMakeCallDataModel{
		PhoneNumber:   "0812345678",
		Variables:     variables,
		Interruptible: true,
		OutboundID:    "test-outbound-123",
		EventID:       "test-event-456",
	})

	assert.NoError(t, err)
	assert.Equal(t, "test-outbound-123", capturedPayload.OutboundID)
	assert.Equal(t, "test-event-456", capturedPayload.EventID)
	assert.Equal(t, "0812345678", capturedPayload.PhoneNumber)
	assert.Equal(t, "35250812345678", capturedPayload.SourcePhone)
	assert.Equal(t, "212", capturedPayload.Speaker)
	assert.Equal(t, "0.1", capturedPayload.FalseSilenceSec)
	assert.Equal(t, "True", capturedPayload.Interruptible)

	// Verify buildFlow carries the new debt-collection variables through as-is.
	assert.Contains(t, capturedPayload.Flow, "สมชาย")
	assert.Contains(t, capturedPayload.Flow, "Toyota Vios กข1234")
	assert.Contains(t, capturedPayload.Flow, "1500.5")
	assert.Contains(t, capturedPayload.Flow, "<!total_interest|120.25!>")
	assert.Contains(t, capturedPayload.Flow, "<!overdue_installment|3!>")
}

func TestVoicebotMakeCallService_ClientError(t *testing.T) {
	mockClient := &mockOutboundBotnoiClient{
		MakeCallFunc: func(payload entities.OutboundBotnoiDataModel) error {
			return errors.New("network timeout")
		},
	}

	svc := &voicebotMakeCallService{
		outboutClient: mockClient,
	}

	err := svc.MakeCall(entities.VoicebotMakeCallDataModel{
		PhoneNumber: "0812345678",
		Variables: map[string]any{
			"name":               "สมชาย",
			"outstanding_amount": "1500.50",
			"due_date":           "2026-06-20",
			"policy_no":          "987654",
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make call: network timeout")
}
