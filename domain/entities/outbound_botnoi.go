package entities

type OutboundBotnoiDataModel struct {
	OutboundID       string `json:"outbound_id,omitempty"`
	EventID          string `json:"event_id,omitempty"`
	PhoneNumber      string `json:"phonenumber,omitempty"`
	Flow             string `json:"flow,omitempty"`
	SourcePhone      string `json:"sourcephone,omitempty"`
	Speaker          string `json:"speaker,omitempty"`
	Language         string `json:"language,omitempty"`
	AgentPhoneNumber string `json:"agent_phone_number,omitempty"`
	Speed            string `json:"speed,omitempty"`
	TTS              string `json:"tts,omitempty"`
	BotID            string `json:"bot_id,omitempty"`
	ASRProvider      string `json:"asr_provider,omitempty"`
	ASRLanguageCode  string `json:"asr_language_code,omitempty"`
	ASRTimeout       int    `json:"asr_timeout,omitempty"`
	FalseTimeoutSec  string `json:"false_timeout_sec,omitempty"`
	FalseSilenceSec  string `json:"false_silence_sec,omitempty"`
	TrueSilenceSec   string `json:"true_silence_sec,omitempty"`
	Interruptible    string `json:"interruptible,omitempty"`
}
