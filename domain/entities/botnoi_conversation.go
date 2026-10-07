package entities

// BotnoiConversationLink ties a Botnoi conversation log (by its conversation id)
// to the call record and debtor it belongs to, so the History tab can show who
// was called instead of a raw id.
type BotnoiConversationLink struct {
	ConversationID string `json:"conversation_id"`
	CallRecordID   string `json:"call_record_id"`
	DebtorName     string `json:"debtor_name"`
	DebtorPhone    string `json:"debtor_phone"`
	Status         string `json:"status"`
	CallOutcome    string `json:"call_outcome"`
	AICategory     string `json:"ai_category"`
	CallDuration   int    `json:"call_duration"`
}
