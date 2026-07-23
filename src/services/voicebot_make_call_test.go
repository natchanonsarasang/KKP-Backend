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

func TestSplitCarDetail(t *testing.T) {
	cases := []struct {
		name         string
		raw          string
		wantPlate    string
		wantProvince string
	}{
		{"plate and province", "ฅฆ 9091 ประจวบคีรีขันธ์", "ฅฆ 9091", "ประจวบคีรีขันธ์"},
		{"new-format plate", "1กก 1234 เชียงใหม่", "1กก 1234", "เชียงใหม่"},
		{"extra spaces", "  ฅฆ 9091   ประจวบคีรีขันธ์  ", "ฅฆ 9091", "ประจวบคีรีขันธ์"},
		{"no province", "ฅฆ 9091", "ฅฆ 9091", ""},
		{"no digits", "รถเก๋ง", "รถเก๋ง", ""},
		{"empty", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plate, province := splitCarDetail(tc.raw)
			assert.Equal(t, tc.wantPlate, plate)
			assert.Equal(t, tc.wantProvince, province)
		})
	}
}

// When car_detail carries a province, the built flow must expose the plate and
// province as two separate variables.
func TestVoicebotMakeCall_SplitsCarDetailIntoProvince(t *testing.T) {
	var capturedPayload entities.OutboundBotnoiDataModel
	mockClient := &mockOutboundBotnoiClient{
		MakeCallFunc: func(payload entities.OutboundBotnoiDataModel) error {
			capturedPayload = payload
			return nil
		},
	}
	svc := &voicebotMakeCallService{outboutClient: mockClient}

	err := svc.MakeCall(entities.VoicebotMakeCallDataModel{
		PhoneNumber: "0812345678",
		Variables: map[string]any{
			"car_detail": "ฅฆ 9091 ประจวบคีรีขันธ์",
		},
	})

	assert.NoError(t, err)
	assert.Contains(t, capturedPayload.Flow, "<!car_detail|ฅฆ 9091!>")
	assert.Contains(t, capturedPayload.Flow, "<!province|ประจวบคีรีขันธ์!>")
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
