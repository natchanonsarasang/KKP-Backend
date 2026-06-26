package entities

type OutboundBotnoiDataModel struct {
	OutboundID       string `json:"outbound_id"`
	EventID          string `json:"event_id"`
	PhoneNumber      string `json:"phonenumber"`
	Flow             string `json:"flow"`
	SourcePhone      string `json:"sourcephone"`
	Speaker          string `json:"speaker"`
	Language         string `json:"language"`
	AgentPhoneNumber string `json:"agent_phone_number"`
	Speed            string `json:"speed"`
	TTS              string `json:"tts"`
	BotID            string `json:"bot_id"`
	ASRProvider      string `json:"asr_provider"`
	ASRLanguageCode  string `json:"asr_language_code"`
	ASRTimeout       int    `json:"asr_timeout"`
	FalseTimeoutSec  string `json:"false_timeout_sec"`
	FalseSilenceSec  string `json:"false_silence_sec"`
	TrueSilenceSec   string `json:"true_silence_sec"`
	Interruptible    string `json:"interruptible"`
}
