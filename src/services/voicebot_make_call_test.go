package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockOutboundBotnoiClient struct {
	MakeCallFunc       func(payload entities.OutboundBotnoiDataModel) error
	CreateBatchFunc    func(phoneNumber, batchName string) (string, error)
	GetBatchStatusFunc func(batchID string) (*entities.OutboundBatchStatusResponse, error)
}

func (m *mockOutboundBotnoiClient) MakeCall(payload entities.OutboundBotnoiDataModel) error {
	if m.MakeCallFunc != nil {
		return m.MakeCallFunc(payload)
	}
	return nil
}

func (m *mockOutboundBotnoiClient) CreateBatch(phoneNumber, batchName string) (string, error) {
	if m.CreateBatchFunc != nil {
		return m.CreateBatchFunc(phoneNumber, batchName)
	}
	return "batch-" + batchName, nil
}

func (m *mockOutboundBotnoiClient) GetBatchStatus(batchID string) (*entities.OutboundBatchStatusResponse, error) {
	if m.GetBatchStatusFunc != nil {
		return m.GetBatchStatusFunc(batchID)
	}
	return &entities.OutboundBatchStatusResponse{}, nil
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

	// Case 2: Nil variables map is fine — V2 no longer sends variables. The call
	// proceeds and a default outbound_id is generated when none is supplied.
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
	assert.Equal(t, "0909722021", capturedPayload.TelephoneNumber)
	assert.True(t, strings.HasPrefix(capturedPayload.OutboundID, "outbound_"))
}

func TestVoicebotMakeCallService_MakeCall(t *testing.T) {
	t.Setenv("OUTBOUND_AGENT_NAME", "collector-agent")

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

	err := svc.MakeCall(entities.VoicebotMakeCallDataModel{
		PhoneNumber: "0812345678",
		OutboundID:  "test-outbound-123",
	})

	assert.NoError(t, err)
	// V2 payload carries exactly three fields.
	assert.Equal(t, "0812345678", capturedPayload.TelephoneNumber)
	assert.Equal(t, "collector-agent", capturedPayload.AgentName)
	assert.Equal(t, "test-outbound-123", capturedPayload.OutboundID)
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
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make call: network timeout")
}
