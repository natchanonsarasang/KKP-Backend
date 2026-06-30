package entities

// WebhookPayload represents the JSON payload received from the webhook.
type WebhookPayload struct {
	OutboundID       string      `json:"outbound_id,omitempty"`
	CallID           string      `json:"call_id,omitempty"`
	Status           string      `json:"status,omitempty"`
	Action           string      `json:"action,omitempty"`
	ConversationLog  string      `json:"conversation_log,omitempty"`
	AudioURL         string      `json:"audio_url,omitempty"`
	Duration         interface{} `json:"duration,omitempty"` // Can be string or number
	CallDuration     interface{} `json:"call_duration,omitempty"`
	PhoneNumber      string      `json:"phone_number,omitempty"`
	AppointmentDate  string      `json:"appointment_date,omitempty"`
	AppointmentTime  string      `json:"appointment_time,omitempty"`
	LastAMDStatus    string      `json:"last_amd_status,omitempty"`
}

