package services

import (
	"errors"
	"fmt"
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/client"
	"os"
	"time"
)

type voicebotMakeCallService struct {
	outboutClient client.IOutboundBotnoiClient
}

type IVoicebotMakeCallService interface {
	MakeCall(data entities.VoicebotMakeCallDataModel) error
}

func NewVoicebotMakeCallService() IVoicebotMakeCallService {
	return &voicebotMakeCallService{
		outboutClient: client.NewOutboundBotnoiClient("", "", ""),
	}
}

func (sv *voicebotMakeCallService) MakeCall(data entities.VoicebotMakeCallDataModel) error {

	if err := validateVoicebotMakeCall(data); err != nil {
		return err
	}

	if data.OutboundID == "" {
		data.OutboundID = fmt.Sprintf("outbound_%d", time.Now().UnixMilli())
	}

	// V2 contract: only the number to dial, the pre-configured agent, and the
	// correlation id. The agent (voicebot persona/script) is selected by name and
	// configured on the Botnoi side, so no flow/TTS/ASR fields are sent.
	payload := entities.OutboundBotnoiDataModel{
		TelephoneNumber: data.PhoneNumber,
		AgentName:       os.Getenv("OUTBOUND_AGENT_NAME"),
		OutboundID:      data.OutboundID,
	}

	err := sv.outboutClient.MakeCall(payload)
	if err != nil {
		return fmt.Errorf("failed to make call: %w", err)
	}

	return nil
}

// function Helper
func validateVoicebotMakeCall(data entities.VoicebotMakeCallDataModel) error {
	if data.PhoneNumber == "" {
		return errors.New("phone_number is required")
	}
	return nil
}

func getStringVal(m map[string]any, key string) string {
	val, ok := m[key]
	if !ok || val == nil {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val)
	}
	return str
}
