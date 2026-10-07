package entities

// WebhookPayload is the JSON the Botnoi provisioning voicebot posts back once a
// call finishes. There is no outbound/call id to correlate on — the call is
// matched to its originating call_record by CallerNumber (the debtor's number).
//
// Status is the call outcome; known values are "completed" and "canceled", and
// more may appear in future. Only "completed" counts as a picked-up/successful
// call — everything else (canceled, failed, ...) is treated as not picked up.
type WebhookPayload struct {
	ConversationID  string `json:"conversation_id,omitempty"`
	AgentID         string `json:"agent_id,omitempty"`
	Status          string `json:"status,omitempty"`
	StartTime       string `json:"start_time,omitempty"`
	EndTime         string `json:"end_time,omitempty"`
	Duration        int    `json:"duration,omitempty"`
	ConversationLog string `json:"conversation_log,omitempty"`
	CallerNumber    string `json:"callerNumber,omitempty"`
}
