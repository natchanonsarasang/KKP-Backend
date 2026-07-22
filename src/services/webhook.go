package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-fiber-template/domain/entities"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type ClassifyResult struct {
	StatusID   int     `json:"status_id"`
	StatusName string  `json:"status_name"`
	Category   string  `json:"category"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}

var CONVERSATION_CATEGORIES = []struct {
	ID    int
	Name  string
	Thai  string
	Group string
}{
	{1, "Convenient to Pay", "สะดวกจ่าย", "main"},
	{2, "Not Convenient to Pay", "ไม่สะดวกจ่าย", "main"},
	{3, "Not Convenient to Talk", "ไม่สะดวกคุย", "main"},
	{4, "Silent", "เงียบ", "main"},
	{5, "Off Topic", "พูดเรื่องอื่น นอกเรื่อง", "main"},
	{6, "Wrong Number", "โทรผิด", "main"},
	{7, "Not Reached", "ติดต่อไม่ได้", "main"},
}

type IWebhookService interface {
	ProcessWebhook(payload entities.WebhookPayload) error
}

type webhookService struct {
	CallRecordsService  ICallRecordsService
	DebtorService       IDebtorsService
	CallListItemService ICallListItemsService
	CallAttemptService  ICallAttemptsService
	CallSessionService  ICallSessionsService
	CallProcessService  ICallProcessService
}

func NewWebhookService(
	callRecords ICallRecordsService,
	debtors IDebtorsService,
	items ICallListItemsService,
	attempts ICallAttemptsService,
	sessions ICallSessionsService,
	callProcess ICallProcessService,
) IWebhookService {
	return &webhookService{
		CallRecordsService:  callRecords,
		DebtorService:       debtors,
		CallListItemService: items,
		CallAttemptService:  attempts,
		CallSessionService:  sessions,
		CallProcessService:  callProcess,
	}
}

func (s *webhookService) ProcessWebhook(payload entities.WebhookPayload) error {
	// Extract fields
	callID := payload.OutboundID
	status := payload.Status
	action := payload.Action
	conversationLog := payload.ConversationLog
	audioURL := payload.AudioURL
	phoneNumber := ""

	// Extract phone number from audio_url if missing (format: ..._PHONE.wav)
	if phoneNumber == "" && audioURL != "" {
		re := regexp.MustCompile(`_(\d+)\.wav`)
		match := re.FindStringSubmatch(audioURL)
		if len(match) > 1 {
			phoneNumber = match[1]
		}
	}

	if callID == "" && phoneNumber == "" {
		log.Warnf("[Webhook] received with no identifiable data (status=%q action=%q): %+v", status, action, payload)
		return nil
	}

	// tag keys every log line for this webhook to the call it belongs to, so a
	// single call can be traced end-to-end across the noisy webhook stream.
	tag := fmt.Sprintf("[Webhook %s/%s]", callID, phoneNumber)
	log.Infof("%s received: status=%q action=%q amd=%q audio=%t log=%t", tag, status, action, payload.LastAMDStatus, audioURL != "", conversationLog != "")

	// Dump the full payload Botnoi sent us so the raw inbound data is always
	// visible in the log for debugging.
	if raw, err := json.Marshal(payload); err == nil {
		log.Infof("%s payload: %s", tag, string(raw))
	} else {
		log.Infof("%s payload (struct): %+v", tag, payload)
	}

	// Check if user actually spoke
	userParts := strings.Split(conversationLog, "User:")
	hasUserSpoken := false
	if len(userParts) > 1 {
		for i := 1; i < len(userParts); i++ {
			trimmed := strings.TrimSpace(userParts[i])
			if len(trimmed) > 0 && !strings.Contains(strings.ToUpper(trimmed), "TIMEOUT") {
				hasUserSpoken = true
				break
			}
		}
	}
	isSilence := len(userParts) > 1 && !hasUserSpoken

	// Map status
	rawStatus := strings.ToLower(status)
	var mappedStatus entities.CallStatus = entities.StatusFailed

	switch {
	case s.contains([]string{"confirm", "yes"}, strings.ToLower(action)):
		mappedStatus = entities.StatusConfirmed
	case s.contains([]string{"decline", "no"}, strings.ToLower(action)):
		mappedStatus = entities.StatusDeclined
	case strings.ToLower(action) == "unknown":
		mappedStatus = entities.StatusNoResponse
	case s.contains([]string{"hanged_up", "hangup", "hung_up"}, rawStatus):
		mappedStatus = entities.StatusHangedUp
	case rawStatus == "completed":
		if hasUserSpoken || isSilence {
			mappedStatus = entities.StatusCompleted
		} else {
			mappedStatus = entities.StatusNoAnswer
		}
	case rawStatus == "no answer" || rawStatus == "no_answer":
		mappedStatus = entities.StatusNoAnswer
	case rawStatus == "busy":
		mappedStatus = entities.StatusBusy
	case rawStatus == "failed" || rawStatus == "error":
		mappedStatus = entities.StatusFailed
	case rawStatus == "rejected":
		mappedStatus = entities.StatusRejected
	case rawStatus == "voicemail":
		mappedStatus = entities.StatusVoicemail
	}

	// Reclassify as "Not Convenient" if needed
	if mappedStatus == entities.StatusNoAnswer && conversationLog != "" {
		if s.askedAboutCallback(conversationLog) {
			mappedStatus = entities.StatusNotConvenient
		}
	}

	// StatusHangedUp means the debtor answered, talked, then hung up — that is a
	// pickup. (A "rejected" status is the caller rejecting the incoming call and
	// stays not-picked-up.) Include it so hanged_up calls, which often arrive with
	// no conversation_log, are still counted as picked up.
	pickedUp := hasUserSpoken || isSilence || s.contains([]string{string(entities.StatusConfirmed), string(entities.StatusDeclined), string(entities.StatusNoResponse), string(entities.StatusCompleted), string(entities.StatusHangedUp)}, string(mappedStatus))

	var finalStatus string
	if mappedStatus == entities.StatusHangedUp {
		finalStatus = "failed"
	} else if pickedUp {
		finalStatus = "success"
	} else {
		finalStatus = "failed"
	}

	outcomeMap := map[entities.CallStatus]string{
		entities.StatusConfirmed:     "Confirmed",
		entities.StatusDeclined:      "Declined",
		entities.StatusNoResponse:    "No Response",
		entities.StatusNoAnswer:      "No Answer",
		entities.StatusCompleted:     "Completed",
		entities.StatusFailed:        "Failed",
		entities.StatusBusy:          "Busy",
		entities.StatusRejected:      "Rejected",
		entities.StatusVoicemail:     "Voicemail",
		entities.StatusHangedUp:      "Hangup",
		entities.StatusNotConvenient: "Not Convenient",
	}
	callOutcome := outcomeMap[mappedStatus]
	if callOutcome == "" {
		callOutcome = "Unknown"
	}

	log.Infof("%s classified: mappedStatus=%s finalStatus=%s pickedUp=%t outcome=%q", tag, mappedStatus, finalStatus, pickedUp, callOutcome)

	// --- AI Categorization ---
	aiResult := s.classifyCall(payload, conversationLog)
	aiCategory := aiResult.Category
	aiReason := aiResult.Reason
	aiConfidence := aiResult.Confidence
	log.Infof("%s ai category=%q reason=%q confidence=%.2f", tag, aiCategory, aiReason, aiConfidence)

	// Resolve Owner (UserID, WorkspaceID)
	var resolvedUserID, resolvedWorkspaceID string

	// 1. Try from CallRecord
	var callRecord *entities.CallRecordDataModel
	if callID != "" {
		records, err := s.CallRecordsService.GetAllCallRecords(entities.CallRecordFilter{BotnoiCallID: callID})
		if err != nil {
			log.Errorf("%s lookup call_record by id failed: %v", tag, err)
		}
		if records != nil && len(*records) > 0 {
			callRecord = &(*records)[0]
			resolvedUserID = callRecord.UserID
			resolvedWorkspaceID = callRecord.WorkspaceID
			if phoneNumber == "" {
				phoneNumber = callRecord.PhoneNumber
			}
		} else {
			log.Warnf("%s no call_record matched call id %q", tag, callID)
		}
	}

	// 2. Fallback: Try from Debtor (WorkspaceID resolve)
	if resolvedWorkspaceID == "" && phoneNumber != "" {
		debtor, err := s.DebtorService.GetDebtorByPhoneNumber(phoneNumber)
		if err != nil {
			log.Errorf("%s lookup debtor by phone failed: %v", tag, err)
		}
		if debtor != nil {
			resolvedUserID = debtor.UserID
			resolvedWorkspaceID = debtor.WorkspaceID
		}
	}

	if resolvedWorkspaceID == "" {
		log.Warnf("%s could not resolve owner (workspace/user) — stats and session advance will be skipped", tag)
	} else {
		log.Infof("%s resolved owner: workspace=%s user=%s", tag, resolvedWorkspaceID, resolvedUserID)
	}

	// Update Call Record and related entities
	if callRecord != nil {
		var duration int
		if payload.Duration != nil {
			duration = s.toInt(payload.Duration)
		}

		callRecord.Status = mappedStatus
		var resultData interface{} = payload
		callRecord.ResultData = &resultData
		callRecord.CallDuration = duration
		callRecord.AppointmentDate = payload.AppointmentDate
		callRecord.AppointmentTime = payload.AppointmentTime
		callRecord.UpdatedAt = time.Now().UTC()

		if err := s.CallRecordsService.UpdateCallRecord(callRecord.ID, *callRecord); err != nil {
			log.Errorf("%s update call_record %s failed: %v", tag, callRecord.ID, err)
		} else {
			log.Infof("%s call_record %s updated: status=%s duration=%ds", tag, callRecord.ID, mappedStatus, duration)
		}

		// Update Call List Items
		items, itemsErr := s.CallListItemService.GetCallListItemsByWorkspace(callRecord.WorkspaceID)
		itemCount := 0
		if items != nil {
			itemCount = len(*items)
		}
		log.Infof("%s [DEBUG-item] lookup workspace=%q err=%v count=%d targetCallRecordID=%q", tag, callRecord.WorkspaceID, itemsErr, itemCount, callRecord.ID)
		matched := false
		if items != nil {
			for _, item := range *items {
				log.Infof("%s [DEBUG-item] candidate item=%q status=%q call_record_id=%q match=%t", tag, item.ID, item.Status, item.CallRecordID, item.CallRecordID == callRecord.ID)
				if item.CallRecordID == callRecord.ID {
					matched = true
					item.Status = finalStatus
					item.CallOutcome = callOutcome
					item.PickedUp = &pickedUp
					item.AICategory = aiCategory
					item.AIReason = aiReason
					item.AIConfidence = aiConfidence
					item.NextRetryAt = nil
					notesObj := map[string]string{
						"audio_url":        audioURL,
						"conversation_log": conversationLog,
					}
					notesJSON, _ := json.Marshal(notesObj)
					item.Notes = string(notesJSON)
					item.UpdatedAt = time.Now().UTC()
					if uerr := s.CallListItemService.UpdateCallListItem(item.ID, item); uerr != nil {
						log.Errorf("%s [DEBUG-item] update item %s failed: %v", tag, item.ID, uerr)
					} else {
						log.Infof("%s [DEBUG-item] item %s updated -> status=%s outcome=%q", tag, item.ID, finalStatus, callOutcome)
					}

					// Update Call Attempt
					attempts, _ := s.CallAttemptService.GetAttemptsByWorkspace(callRecord.WorkspaceID)
					updated := false
					if attempts != nil {
						for _, attempt := range *attempts {
							// Find the "calling" attempt for this item
							if attempt.CallListItemID == item.ID && attempt.Status == "calling" {
								attempt.Status = finalStatus
								attempt.CallOutcome = callOutcome
								attempt.PickedUp = &pickedUp
								attempt.AiCategory = aiCategory
								attempt.AiReason = aiReason
								attempt.AiConfidence = aiConfidence
								attempt.ConversationLog = conversationLog
								attempt.AudioURL = audioURL
								attempt.CallDuration = duration
								attempt.CallRecordID = callRecord.ID
								attempt.UpdatedAt = time.Now().UTC()
								s.CallAttemptService.UpdateAttempt(attempt.ID, attempt)
								updated = true
								break
							}
						}
					}

					// Fallback: If no "calling" attempt was found, insert a new one
					if !updated {
						newAttempt := entities.CallAttemptModel{
							UserID:          resolvedUserID,
							CallListItemID:  item.ID,
							CallRecordID:    callRecord.ID,
							WorkspaceID:     resolvedWorkspaceID,
							Status:          finalStatus,
							CallOutcome:     callOutcome,
							PickedUp:        &pickedUp,
							AiCategory:      aiCategory,
							AiReason:        aiReason,
							AiConfidence:    aiConfidence,
							ConversationLog: conversationLog,
							AudioURL:        audioURL,
							CallDuration:    duration,
						}
						s.CallAttemptService.CreateAttempt(newAttempt)
					}

					// Create a "finished" attempt document recording the webhook-confirmed outcome.
					nowAttempt := time.Now().UTC()
					attemptNumber := 1
					if attempts != nil {
						for _, a := range *attempts {
							if a.CallListItemID == item.ID {
								attemptNumber++
							}
						}
					}
					s.CallAttemptService.CreateAttempt(entities.CallAttemptModel{
						ID:              uuid.NewString(),
						UserID:          resolvedUserID,
						WorkspaceID:     resolvedWorkspaceID,
						CallListItemID:  item.ID,
						CallRecordID:    callRecord.ID,
						AttemptNumber:   attemptNumber,
						Status:          "finished",
						CallOutcome:     callOutcome,
						PickedUp:        &pickedUp,
						AiCategory:      aiCategory,
						AiReason:        aiReason,
						AiConfidence:    aiConfidence,
						ConversationLog: conversationLog,
						AudioURL:        audioURL,
						CallDuration:    duration,
						ErrorReason:     "",
						CreatedAt:       nowAttempt,
						UpdatedAt:       nowAttempt,
					})
				}
			}
		}
		if !matched {
			log.Warnf("%s [DEBUG-item] NO call_list_item matched call_record_id=%q among %d workspace items", tag, callRecord.ID, itemCount)
		}
	}

	// Update Debtor Stats
	if phoneNumber != "" && resolvedWorkspaceID != "" {
		debtors, _ := s.DebtorService.GetDebtorsByWorkspace(resolvedWorkspaceID)
		if debtors != nil {
			for _, debtor := range *debtors {
				if debtor.PhoneNumber == phoneNumber {
					nowTime := time.Now().UTC()
					debtor.LastContactAt = &nowTime
					debtor.UpdatedAt = nowTime
					debtor.CallOutcome = string(mappedStatus)
					debtor.CallAnswered = &pickedUp
					debtor.ContactAttempts++

					if pickedUp {
						debtor.PickedUpCount++
						debtor.SuccessfulContacts++
					} else {
						debtor.NotPickedUpCount++
					}

					if mappedStatus == entities.StatusConfirmed {
						debtor.LastResponse = "accept"
					} else if mappedStatus == entities.StatusDeclined {
						debtor.LastResponse = "reject"
					} else if mappedStatus == entities.StatusNoResponse || (mappedStatus == entities.StatusCompleted && pickedUp) {
						debtor.LastResponse = "unknown"
					}

					// Extract callback date from conversation log
					dateCon := s.extractCallbackDate(conversationLog)
					debtor.DateCon = dateCon

					s.DebtorService.UpdateDebtor(debtor.ID, debtor)
					break
				}
			}
		}
	}

	// Update active session stats and trigger next call
	if resolvedWorkspaceID != "" {
		sessions, _ := s.CallSessionService.GetCallSessions(entities.CallSessionFilter{WorkspaceID: resolvedWorkspaceID, Status: "running"})
		if sessions != nil {
			for _, session := range *sessions {
				if session.Status == "running" && session.WorkspaceID == resolvedWorkspaceID {
					if finalStatus == "success" {
						session.CompletedCalls++
						if mappedStatus == entities.StatusConfirmed {
							session.ConfirmedCalls++
						}
					} else if finalStatus == "failed" {
						session.FailedCalls++
					}
					session.UpdatedAt = time.Now().UTC()
					if err := s.CallSessionService.UpdateCallSession(session.ID, session); err != nil {
						log.Errorf("%s update session %s failed: %v", tag, session.ID, err)
					} else {
						log.Infof("%s session %s advanced: completed=%d confirmed=%d failed=%d", tag, session.ID, session.CompletedCalls, session.ConfirmedCalls, session.FailedCalls)
					}

					// Trigger the next batch via the call-process service directly.
					// Run in a goroutine so the webhook response is not blocked by
					// the (potentially long-running, recursive) session processing.
					sessionID := session.ID
					go func() {
						if err := s.CallProcessService.ProcessSession(sessionID); err != nil {
							log.Errorf("%s ProcessSession %s failed: %v", tag, sessionID, err)
						}
					}()
				}
			}
		}
	}

	log.Infof("%s done", tag)
	return nil
}

func (s *webhookService) classifyCall(payload entities.WebhookPayload, logText string) ClassifyResult {
	if os.Getenv("GROQ_API_KEY") == "" || logText == "" || len(logText) < 5 {
		return ClassifyResult{Category: "Not Reached", Reason: "No log or API key missing", Confidence: 0}
	}

	// System-level status map
	systemStatusMap := map[string]string{
		"no_answer":   "Not Reached",
		"no answer":   "Not Reached",
		"unreachable": "Not Reached",
		"rejected":    "Not Reached",
		"busy":        "Not Reached",
		"voicemail":   "Not Reached",
		"failed":      "Not Reached",
	}
	rawStatus := strings.ToLower(strings.TrimSpace(payload.Status))
	if cat, ok := systemStatusMap[rawStatus]; ok {
		return ClassifyResult{Category: cat, Reason: "System status: " + rawStatus, Confidence: 1}
	}

	// Rule-based silence detection
	userTurns := strings.Split(logText, "User:")
	if len(userTurns) > 1 {
		hasRealSpeech := false
		for i := 1; i < len(userTurns); i++ {
			trimmed := strings.TrimSpace(strings.ToUpper(userTurns[i]))
			if len(trimmed) > 0 && !strings.Contains(trimmed, "TIMEOUT") {
				hasRealSpeech = true
				break
			}
		}
		if !hasRealSpeech {
			return ClassifyResult{Category: "Silent", Reason: "Customer picked up but remained silent (ASR TIMEOUT)", Confidence: 1}
		}
	}

	// Rule-based audio-quality detection
	audioQualityPatterns := []string{
		"can't hear", "cannot hear", "hard to hear", "loud noise", "too noisy",
		"background noise", "unclear audio", "audio is unclear", "breaking up",
		"ไม่ได้ยิน", "เสียงไม่ชัด", "เสียงดัง", "เสียงรบกวน", "เสียงแทรก",
	}
	for _, p := range audioQualityPatterns {
		if strings.Contains(strings.ToLower(logText), p) {
			return ClassifyResult{Category: "Not Convenient to Talk", Reason: "Detected audio-quality keywords", Confidence: 1}
		}
	}

	// AI Prompt
	categoryList := ""
	for _, c := range CONVERSATION_CATEGORIES {
		categoryList += fmt.Sprintf("%d. %s (%s) [%s]\n", c.ID, c.Name, c.Thai, c.Group)
	}

	systemPrompt := `You classify Thai debt-collection call transcripts. Return STRICT JSON only.
Choose exactly ONE category (use the EXACT English label) from this list:
` + categoryList + `

CATEGORY DEFINITIONS
- Convenient to Pay      → Customer is able/willing to pay: confirms payment, agrees to a payment date or plan, says they will pay ("จ่ายได้", "โอนให้", "พรุ่งนี้จ่าย"), states it is already paid, or otherwise acknowledges the debt with clear intent to pay.
- Not Convenient to Pay  → Customer cannot or will not pay now: says they have no money / cannot afford it, refuses to pay, denies the debt, or asks to restructure / defer / pay in installments because paying now is not possible.
- Not Convenient to Talk → Customer picked up but it is not a convenient time to talk: says they are busy, asks to be called back later, or cannot hear clearly due to audio problems.
- Silent                 → Customer picked up but remained silent throughout with no meaningful verbal response.
- Off Topic              → Customer kept talking about unrelated topics / went off-topic with no resolution.
- Wrong Number           → Customer says this is the wrong number / not the intended person / not the policyholder.
- Not Reached            → Customer could not actually be contacted (no answer, line dead, voicemail, unreachable, or hung up before any meaningful exchange).

CLASSIFICATION RULES
1. Prefer a payment outcome (Convenient to Pay / Not Convenient to Pay) whenever the customer engaged with the debt topic.
2. Use Not Convenient to Talk only when the customer engaged but could not talk now and gave no payment answer.
3. Use Wrong Number / Silent / Off Topic / Not Reached only when no payment outcome applies.

Output format (STRICT JSON):
{
  "status_name": "<exact English label from the list>",
  "confidence": <number between 0 and 1>,
  "reason": "<short explanation>"
}`

	rawContent, err := s.callGroqChat("llama-3.3-70b-versatile", systemPrompt, `conversation_log:\n"""`+logText+`"""`)
	if err != nil {
		log.Errorf("AI Classify Error: %v", err)
		return ClassifyResult{Category: "Not Reached", Reason: "AI request failed", Confidence: 0}
	}

	var content struct {
		StatusName string  `json:"status_name"`
		Reason     string  `json:"reason"`
		Confidence float64 `json:"confidence"`
	}
	json.Unmarshal([]byte(rawContent), &content)

	aiName := strings.ToLower(strings.TrimSpace(content.StatusName))

	for _, c := range CONVERSATION_CATEGORIES {
		if strings.ToLower(c.Name) == aiName {
			return ClassifyResult{
				StatusID:   c.ID,
				StatusName: c.Name,
				Category:   c.Name,
				Reason:     content.Reason,
				Confidence: content.Confidence,
			}
		}
	}

	return ClassifyResult{Category: "Not Reached", Reason: "Defaulted or unmatched", Confidence: 0}
}

// callGroqChat sends an OpenAI-compatible chat-completion request to Groq's free
// API and returns the assistant message content. All calls request strict JSON
// output at temperature 0. Requires GROQ_API_KEY.
func (s *webhookService) callGroqChat(model, systemPrompt, userContent string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	type aiRequest struct {
		Model          string                 `json:"model"`
		Messages       []map[string]string    `json:"messages"`
		ResponseFormat map[string]interface{} `json:"response_format"`
		Temperature    float64                `json:"temperature"`
	}

	reqBody := aiRequest{
		Model: model,
		Messages: []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userContent},
		},
		ResponseFormat: map[string]interface{}{"type": "json_object"},
		Temperature:    0,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq status %d: %s", resp.StatusCode, string(body))
	}

	var aiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return "", err
	}
	if len(aiResponse.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}
	return aiResponse.Choices[0].Message.Content, nil
}

func (s *webhookService) extractCallbackDate(logText string) string {
	if os.Getenv("GROQ_API_KEY") == "" || logText == "" || len(logText) < 5 {
		return ""
	}

	refIso := time.Now().Format("2006-01-02")
	re := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	match := re.FindStringSubmatch(logText)
	if len(match) > 0 {
		refIso = match[0]
	}

	systemPrompt := `You extract a callback date from a Thai debt-collection call transcript.
Reference (call) date in Asia/Bangkok timezone: ` + refIso + `
Rules:
- Exact date stated → return it. Subtract 543 if Buddhist Era.
- "พรุ่งนี้" → ref + 1 day
- "มะรืน" → ref + 2 days
- "อีก X วัน" → ref + X days
- "สัปดาห์หน้า" → ref + 7 days
Return STRICT JSON only: { "date_con": "YYYY-MM-DD" | null }`

	rawContent, err := s.callGroqChat("llama-3.3-70b-versatile", systemPrompt, `conversation_log:\n"""`+logText+`"""`)
	if err != nil {
		log.Errorf("AI Date Extract Error: %v", err)
		return ""
	}

	var content struct {
		DateCon string `json:"date_con"`
	}
	json.Unmarshal([]byte(rawContent), &content)
	return content.DateCon
}

func (s *webhookService) triggerSessionProcessor(sessionID string) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		return
	}

	url := fmt.Sprintf("%s/functions/v1/process-call-session", supabaseURL)
	payload := map[string]string{"session_id": sessionID, "action": "continue"}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (s *webhookService) toInt(v interface{}) int {
	switch i := v.(type) {
	case float64:
		return int(math.Round(i))
	case int:
		return i
	case string:
		var f float64
		fmt.Sscanf(i, "%f", &f)
		return int(math.Round(f))
	default:
		return 0
	}
}

func (s *webhookService) contains(slice []string, val string) bool {
	for _, sliceVal := range slice {
		if sliceVal == val {
			return true
		}
	}
	return false
}

func (s *webhookService) askedAboutCallback(logText string) bool {
	logText = strings.ToLower(logText)
	keywords := []string{
		"convenient", "callback", "call back", "call you back",
		"what day", "what time", "which day", "which time",
		"when would", "when can", "when is", "available",
		"สะดวก", "นัด", "วันไหน", "เวลาไหน", "ติดต่อใหม่", "โทรกลับ",
	}
	for _, k := range keywords {
		if strings.Contains(logText, k) {
			return true
		}
	}
	return false
}
