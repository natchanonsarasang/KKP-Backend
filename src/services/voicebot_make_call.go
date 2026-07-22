package services

import (
	"errors"
	"fmt"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/utils"
	"go-fiber-template/src/client"
	"os"
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
	variables["bot_type"] = "in_init_conversation"
	variables["intent"] = "in_init_conversation"

	 var interruptible string
	 if data.Interruptible {
	 	interruptible = "True"
	 } else {
	 	interruptible = "False"
	 }

	payload := entities.OutboundBotnoiDataModel{
		OutboundID: data.OutboundID,
		Flow: buildFlow(data.OutboundID,
			getStringVal(variables, "name"),
			getStringVal(variables, "car_detail"),
			getStringVal(variables, "total_debt"),
			getStringVal(variables, "total_interest"),
			getStringVal(variables, "total_fine"),
			getStringVal(variables, "overdue_installment"),
			getStringVal(variables, "bot_type"),
			getStringVal(variables, "intent")),
		PhoneNumber: data.PhoneNumber,
		BotID:       "6a06964fb875327d960f05f0",
		BotType:     os.Getenv("BOT_TYPE"),
		Intent:      "in_init_conversation",
		// The partner /outbound contract only requires outbound_id, phonenumber,
		// flow, bot_id. The extra call-config fields below are kept (commented)
		// for future use — re-enable if the partner API stops applying defaults.
		 EventID:          data.EventID,
		 SourcePhone:      "3525" + data.PhoneNumber,
		 Speaker:          "212",
		 Language:         "th",
		 AgentPhoneNumber: "0800000000",
		 Speed:            "1",
		 TTS:              "voicebot-premium",
		 ASRProvider:      "botnoi-th-noise-classifier-C",
		 ASRLanguageCode:  "th",
		 ASRTimeout:       5,
		 FalseTimeoutSec:  "1",
		 FalseSilenceSec:  "0.1",
		 TrueSilenceSec:   "0.25",
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
	carDetail string, totalDebt string, totalInterest string, totalFine string,
	overdueInstallment string, botType string, intent string) string {
	flow := fmt.Sprintf(
		"<!outbound_id|%s!>|||"+
			"<!customer_name|%s!>|||"+
			"<!car_detail|%s!>|||"+
			"<!total_debt|%s!>|||"+
			"<!total_interest|%s!>|||"+
			"<!total_fine|%s!>|||"+
			"<!overdue_installment|%s!>|||"+
			"<!bot_type|%s!>|||"+
			"<!intent|%s!>|||"+
			"{{in_init_conversation}}",
		outboundID, name, carDetail, totalDebt, totalInterest, totalFine, overdueInstallment, botType, intent)

	return flow
}
