package entities

// WebhookPayload represents the JSON payload received from the webhook.
type WebhookPayload struct {
	OutboundID       string      `json:"outbound_id"`
	CallID           string      `json:"call_id"`
	Status           string      `json:"status"`
	Action           string      `json:"action"`
	ConversationLog  string      `json:"conversation_log"`
	AudioURL         string      `json:"audio_url"`
	Duration         interface{} `json:"duration"` // Can be string or number
	CallDuration     interface{} `json:"call_duration"`
	PhoneNumber      string      `json:"phone_number"`
	AppointmentDate  string      `json:"appointment_date"`
	AppointmentTime  string      `json:"appointment_time"`
	Message          string      `json:"message"`
	LastAMDStatus    string      `json:"last_amd_status"`
	Error            string      `json:"error"`
}
