package entities

// DebtorStatusCount is a single (debtor, call-record status) bucket produced by
// the call-stats aggregation. One debtor yields several of these — one per
// distinct status seen across its call_records.
type DebtorStatusCount struct {
	DebtorID string `json:"debtor_id" bson:"debtor_id"`
	Status   string `json:"status" bson:"status"`
	Count    int    `json:"count" bson:"count"`
}

// DebtorCallStats is the per-debtor call summary the debtor list displays. It is
// derived from call_records (the source of truth) rather than the incremental
// counters on the debtor document, so it never drifts.
type DebtorCallStats struct {
	Total       int `json:"total"`
	Confirmed   int `json:"confirmed"`
	Declined    int `json:"declined"`
	NoResponse  int `json:"no_response"`
	PickedUp    int `json:"picked_up"`
	NotPickedUp int `json:"not_picked_up"`
}
