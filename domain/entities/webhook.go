package entities

// WebhookPayload is the JSON the Botnoi provisioning voicebot posts back once a
// call finishes. ConversationID is the call_id Botnoi returned when the call was
// placed (stored as call_record.botnoi_call_id), so it is the primary key we
// correlate the result back to its originating call_record on; CallerNumber (the
// debtor's number) is the fallback when ConversationID is missing/unmatched.
//
// Status is Botnoi's raw call outcome (e.g. "completed", "canceled", "failed").
// It is NOT a reliable "was it answered" signal on its own — Botnoi sends
// "canceled" even for calls that connected and held a full conversation. Whether
// the debtor was actually reached is derived from the transcript/duration in the
// webhook handler, not from Status alone.
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
	// AudioURL is the (pre-signed) recording of the call, present when Botnoi has
	// an audio log for it. Stored on the call_attempt so the UI can play it back.
	AudioURL        string `json:"audio_url,omitempty"`
	CallerNumber    string `json:"callerNumber,omitempty"`
}
