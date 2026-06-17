package services

import (
	"errors"
	"fmt"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/utils"
	"go-fiber-template/src/client"
	"strings"
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
	if data.EventID == "" {
		data.EventID = fmt.Sprintf("event_%d", time.Now().UnixMilli())
	}
	variables := prepareVoicebotVariables(data.Variables)

	var interruptible string
	if data.Interruptible {
		interruptible = "True"
	} else {
		interruptible = "False"
	}

	payload := entities.OutboundBotnoiDataModel{
		OutboundID: data.OutboundID,
		EventID:    data.EventID,
		Flow: buildFlow(data.OutboundID,
			getStringVal(variables, "name"),
			getStringVal(variables, "outstanding_amount"),
			getStringVal(variables, "overdue_installment"),
			getStringVal(variables, "due_date"),
			getStringVal(variables, "policy_no")),
		PhoneNumber:      data.PhoneNumber,
		SourcePhone:      "3525" + data.PhoneNumber,
		Speaker:          "212",
		Language:         "th",
		AgentPhoneNumber: "0800000000",
		Speed:            "1",
		TTS:              "voicebot-premium",
		BotID:            "6a06964fb875327d960f05f0",
		ASRProvider:      "botnoi-aws-th-noise-classifier-v17c",
		ASRLanguageCode:  "th",
		ASRVadRules:      entities.ASRVadRules{FalseTimeoutSec: 1, FalseSilenceSec: 0.1, TrueSilenceSec: 0.25},
		Interruptible:    interruptible,
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
	if data.Variables == nil {
		return errors.New("variables is required")
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

func prepareVoicebotVariables(input map[string]any) map[string]any {
	variables := make(map[string]any)

	for key, value := range input {
		variables[key] = value
	}

	policyNo := getStringVal(variables, "policy_no")
	if policyNo != "" {
		raw := strings.TrimSpace(policyNo)

		if raw != "" {
			variables["policy_no_raw"] = raw
			variables["policy_no"] = utils.ToThaiDigitSpeech(raw)
		}
	}

	date_today := utils.ThaiTodayString()
	variables["date_today"] = date_today

	return variables
}

func buildFlow(outboundID string, name string,
	outstandingAmount string, overdueInstallment string, dueDate string,
	policyNo string) string {
	flow := fmt.Sprintf(
		"<!outbound_id|%s!>|||"+
			"<!name|%s!>|||"+
			"<!outstanding_amount|%s!>|||"+
			"<!overdue_installment|%s!>|||"+
			"<!due_date|%s!>|||"+
			"<!2|2!>|||"+
			"<!policy_no|%s!>|||"+
			"{{Confirm1}}",
		outboundID, name, outstandingAmount, overdueInstallment, dueDate, policyNo)

	return flow
}
