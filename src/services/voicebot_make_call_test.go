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

	// Case 2: Nil variables map
	err = svc.MakeCall(entities.VoicebotMakeCallDataModel{
		PhoneNumber: "0909722021",
		Variables:   nil,
	})
	assert.Error(t, err)
	assert.Equal(t, "variables is required", err.Error())
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
		"name":               "สมชาย",
		"outstanding_amount": 1500.50,
		"due_date":           "2026-06-20",
		"policy_no":          "987654",
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
	assert.Equal(t, "", capturedPayload.FalseSilenceSec)
	assert.Equal(t, "True", capturedPayload.Interruptible)

	// Verify buildFlow content and Thai digits speech transformation for policy_no
	// policy_no "987654" in Thai digit speech:
	// 9 -> เก้า
	// 8 -> แปด
	// 7 -> เจ็ด
	// 6 -> หก
	// 5 -> ห้า
	// 4 -> สี่
	// So policy_no should contain "เก้า แปด เจ็ด หก ห้า สี่"
	assert.Contains(t, capturedPayload.Flow, "เก้า แปด เจ็ด หก ห้า สี่")
	assert.Contains(t, capturedPayload.Flow, "สมชาย")
	assert.Contains(t, capturedPayload.Flow, "1500.5")
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
