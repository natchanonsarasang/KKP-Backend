package entities

// WebhookPayload is the JSON the Botnoi provisioning voicebot posts back once a
// call finishes. ConversationID is the call_id Botnoi returned when the call was
// placed (stored as call_record.botnoi_call_id), so it is the primary key we
// correlate the result back to its originating call_record on; CallerNumber (the
// debtor's number) is the fallback when ConversationID is missing/unmatched.
//
// Status is the call outcome; known values are "completed" and "canceled", and
// more may appear in future. Only "completed" counts as a picked-up/successful
// call — everything else (canceled, failed, ...) is treated as not picked up.
type WebhookPayload struct {
	ConversationID  string `json:"conversation_id,omitempty"`
	AgentID         string `json:"agent_id,omitempty"`
	Status          string `json:"status,omitempty"`
	// Reason is Botnoi's machine-readable failure cause for non-completed calls
	// (e.g. "unreachable", "busy"). Empty for a completed call.
	Reason          string `json:"reason,omitempty"`
	StartTime       string `json:"start_time,omitempty"`
	EndTime         string `json:"end_time,omitempty"`
	Duration        int    `json:"duration,omitempty"`
	ConversationLog string `json:"conversation_log,omitempty"`
	CallerNumber    string `json:"callerNumber,omitempty"`
}
