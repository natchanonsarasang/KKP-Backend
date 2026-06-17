package entities

type VoicebotMakeCallDataModel struct {
	PhoneNumber   string         `json:"phone_number"`
	Variables     map[string]any `json:"variables"`
	Interruptible bool           `json:"interruptible"`
	NextIntent    string         `json:"next_intent"`
	OutboundID    string         `json:"outbound_id"`
	EventID       string         `json:"event_id"`
}
