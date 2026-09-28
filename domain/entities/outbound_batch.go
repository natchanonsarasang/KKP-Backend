package entities

// OutboundBatchCreateRequest is the body for POST /api/outbound/batch. A batch is
// created up-front (one phone number per batch in our flow) and the batch_id it
// returns is then used as the outbound_id when placing the actual call.
type OutboundBatchCreateRequest struct {
	PhoneNumbers []string `json:"phone_numbers"`
	AgentName    string   `json:"agent_name"`
	BatchName    string   `json:"batch_name"`
}

// OutboundBatchCreateResponse is the response from POST /api/outbound/batch.
type OutboundBatchCreateResponse struct {
	Message string `json:"message"`
	BatchID string `json:"batch_id"`
}

// OutboundBatchStatusResponse is the response from GET /api/outbound/batch/{id}.
// While the webhook is not yet available we poll this endpoint to learn a call's
// outcome: `Calls` is empty ([]) until the call has been placed and produces a
// result, then it carries one entry per dialed number.
type OutboundBatchStatusResponse struct {
	Batch OutboundBatchInfo   `json:"batch"`
	Calls []OutboundBatchCall `json:"calls"`
}

type OutboundBatchInfo struct {
	BatchID      string `json:"batch_id"`
	BatchName    string `json:"batch_name"`
	AgentID      string `json:"agent_id"`
	AgentName    string `json:"agent_name"`
	CreatedBy    string `json:"created_by"`
	Status       string `json:"status"`
	TotalNumbers int    `json:"total_numbers"`
	SuccessCalls int    `json:"success_calls"`
	FailedCalls  int    `json:"failed_calls"`
	CreatedAt    string `json:"created_at"`
}

// OutboundBatchCall is one dialed number's result inside a batch. Only call-level
// fields are available here (there is no conversation transcript — that arrives
// via the webhook once it is finished), so classification is call-level only.
type OutboundBatchCall struct {
	SessionID            string `json:"session_id"`
	Status               string `json:"status"`
	UserID               string `json:"user_id"`
	BotID                string `json:"bot_id"`
	PhoneNumber          string `json:"phone_number"`
	AgentName            string `json:"agent_name"`
	CallType             string `json:"call_type"`
	CreatedAt            string `json:"created_at"`
	OutboundID           string `json:"outbound_id"`
	UpdatedAt            string `json:"updated_at"`
	AMDStopReason        string `json:"amd_stop_reason"`
	AMDVoicemailDetected bool   `json:"amd_voicemail_detected"`
	BeepDetected         bool   `json:"beep_detected"`
	CallDuration         string `json:"call_duration"`
	CallStateValue       string `json:"call_state_value"`
	HadAnswered          bool   `json:"had_answered"`
	HangupReason         string `json:"hangup_reason"`
	LastAMDStatus        string `json:"last_amd_status"`
}
