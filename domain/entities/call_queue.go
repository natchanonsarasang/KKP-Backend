package entities

import "time"

// CallQueueModel is one entry in the outbound call queue. Because the V2 Botnoi
// voicebot fetches debtor data mid-call via a "Tools KKP_Data" HTTP call that
// carries NO identifier, we cannot look the debtor up on demand. Instead, when a
// call is placed we push a queue row holding a snapshot of that debtor's
// variables; the KKP_Data endpoint serves the head row, and the webhook deletes
// it (matched by OutboundID) once the call finishes. This only works while calls
// are serialized to one at a time — see ProcessSession forcing maxConcurrent=1.
type CallQueueModel struct {
	ID             string            `bson:"id,omitempty" json:"id"`
	OutboundID     string            `bson:"outbound_id,omitempty" json:"outbound_id"`
	CallListItemID string            `bson:"call_list_item_id,omitempty" json:"call_list_item_id"`
	DebtorID       string            `bson:"debtor_id,omitempty" json:"debtor_id"`
	SessionID      string            `bson:"session_id,omitempty" json:"session_id"`
	WorkspaceID    string            `bson:"workspace_id,omitempty" json:"workspace_id"`
	UserID         string            `bson:"user_id,omitempty" json:"user_id"`
	PhoneNumber    string            `bson:"phone_number,omitempty" json:"phone_number"`
	Variables      map[string]string `bson:"variables,omitempty" json:"variables"`
	CreatedAt      time.Time         `bson:"created_at,omitempty" json:"created_at"`
}
