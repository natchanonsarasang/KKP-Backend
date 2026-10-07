package entities

// OutboundCallRequest is the body for the Botnoi provisioning "place a call"
// endpoint: POST /v1/provisioning/tenants/{tenant_id}/calls. The voicebot now
// receives the debtor's data up-front via Metadata (there is no mid-call
// KKP_Data fetch), and the agent (voicebot persona/script) is selected by
// AgentID, pre-configured on the Botnoi side.
type OutboundCallRequest struct {
	// Destination is the phone number to dial (the debtor's number).
	Destination string `json:"destination"`
	// AgentID selects the pre-configured Botnoi agent (e.g. "agt_d2553cbbe2a2").
	AgentID string `json:"agent_id"`
	// Path is an optional routing hint ("pbx" | "pstn"); omitted when nil.
	Path *string `json:"path,omitempty"`
	// Metadata carries the debtor variables the agent reads during the call
	// (customer_name, car_detail, overdue_installment, total_debt, ...).
	Metadata map[string]any `json:"metadata,omitempty"`
}
