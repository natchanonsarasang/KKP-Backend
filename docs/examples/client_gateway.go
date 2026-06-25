package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL  = "http://localhost:8080"
	jwtToken = "YOUR_JWT_TOKEN"
)

// VoicebotMakeCallDataModel matches database model structure
type VoicebotMakeCallDataModel struct {
	PhoneNumber   string                 `json:"phone_number"`
	Variables     map[string]interface{} `json:"variables"`
	Interruptible bool                   `json:"interruptible"`
	NextIntent    string                 `json:"next_intent"`
	OutboundID    string                 `json:"outbound_id"`
	EventID       string                 `json:"event_id"`
}

// TriggerVoicebotDial manually routes call request through http gateway
func TriggerVoicebotDial(phoneNumber string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload := VoicebotMakeCallDataModel{
		PhoneNumber: phoneNumber,
		Variables: map[string]interface{}{
			"name":       "John Doe",
			"due_amount": "5,500 THB",
		},
		Interruptible: true,
		NextIntent:    "confirm_identity",
		OutboundID:    "btn_out_91283",
		EventID:       "evt_338472",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/api/v1/voicebot/make-call", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set Headers
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Println("Live voicebot make-call trigger sent successfully.")
	return nil
}

func main() {
	err := TriggerVoicebotDial("0812345678")
	if err != nil {
		fmt.Printf("Execution Error: %v\n", err)
	}
}
