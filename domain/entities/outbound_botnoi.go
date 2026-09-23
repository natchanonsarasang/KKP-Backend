package entities

// OutboundBotnoiDataModel is the V2 request body for the Botnoi POST /outbound
// endpoint. The V2 contract only carries three fields — the agent (voicebot
// script/persona) is pre-configured on the Botnoi side and selected by name, so
// the caller no longer sends flow/TTS/ASR configuration.
type OutboundBotnoiDataModel struct {
	TelephoneNumber string `json:"telephone_number"`
	AgentName       string `json:"agent_name"`
	OutboundID      string `json:"outbound_id"`
}
