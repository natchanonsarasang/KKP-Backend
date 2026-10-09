package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-fiber-template/domain/entities"
	"io"
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
	// Extract fields. Botnoi echoes no outbound/call id, so the debtor's number
	// (callerNumber) is the only key we can correlate the call back on.
	status := strings.ToLower(strings.TrimSpace(payload.Status))
	conversationLog := payload.ConversationLog
	phoneNumber := payload.CallerNumber
	duration := payload.Duration
	// audio is fetched via a separate endpoint we don't call yet; keep the notes
	// key present (empty) so the frontend's parsing stays stable.
	audioURL := ""

	if phoneNumber == "" {
		log.Warnf("[Webhook] received with no callerNumber (status=%q): %+v", status, payload)
		return nil
	}

	// tag keys every log line for this webhook to the call it belongs to, so a
	// single call can be traced end-to-end across the noisy webhook stream.
	tag := fmt.Sprintf("[Webhook %s/%s]", payload.ConversationID, phoneNumber)
	log.Infof("%s received: status=%q duration=%d log=%t", tag, status, duration, conversationLog != "")

	// Dump the full payload Botnoi sent us so the raw inbound data is always
	// visible in the log for debugging.
	if raw, err := json.Marshal(payload); err == nil {
		log.Infof("%s payload: %s", tag, string(raw))
	} else {
		log.Infof("%s payload (struct): %+v", tag, payload)
	}

	// Any webhook means the call reached Botnoi and finished, so it is recorded as
	// a completed, picked-up call whatever Botnoi's status says (completed,
	// canceled, no_answer, ...). The raw status is still kept in result_data.
	pickedUp := true
	mappedStatus := entities.StatusCompleted
	finalStatus := "success"
	callOutcome := "Completed"

	log.Infof("%s classified: mappedStatus=%s finalStatus=%s pickedUp=%t outcome=%q", tag, mappedStatus, finalStatus, pickedUp, callOutcome)

	// --- AI Categorization ---
	aiResult := s.classifyCall(payload, conversationLog)
	aiCategory := aiResult.Category
	aiReason := aiResult.Reason
	aiConfidence := aiResult.Confidence

	log.Infof("%s ai category=%q reason=%q confidence=%.2f", tag, aiCategory, aiReason, aiConfidence)

	// Resolve Owner (UserID, WorkspaceID)
	var resolvedUserID, resolvedWorkspaceID string

	// 1. Match the originating call_record by phone number. The dialer leaves it
	// "pending" until this webhook lands; if several pending records share the
	// number (a prior call never resolved), take the most recent.
	var callRecord *entities.CallRecordDataModel
	records, err := s.CallRecordsService.GetAllCallRecords(entities.CallRecordFilter{
		PhoneNumber: phoneNumber,
		Status:      string(entities.StatusPending),
	})
	if err != nil {
		log.Errorf("%s lookup call_record by phone failed: %v", tag, err)
	}
	if records != nil && len(*records) > 0 {
		callRecord = &(*records)[0]
		for i := range *records {
			if (*records)[i].CreatedAt.After(callRecord.CreatedAt) {
				callRecord = &(*records)[i]
			}
		}
		resolvedUserID = callRecord.UserID
		resolvedWorkspaceID = callRecord.WorkspaceID
	} else {
		log.Warnf("%s no pending call_record matched phone %q", tag, phoneNumber)
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
		callRecord.Status = mappedStatus
		var resultData interface{} = payload
		callRecord.ResultData = &resultData
		callRecord.CallDuration = duration
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

					// No explicit confirm/decline signal anymore — derive intent from
					// the AI category, defaulting any other picked-up call to "unknown".
					if aiCategory == "Convenient to Pay" {
						debtor.LastResponse = "accept"
					} else if aiCategory == "Not Convenient to Pay" {
						debtor.LastResponse = "reject"
					} else if pickedUp {
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
						// Without an explicit confirm/decline signal, treat the AI
						// "Convenient to Pay" category as a confirmed-to-pay call.
						if aiCategory == "Convenient to Pay" {
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

	rawContent, err := s.callGroqChat(groqModel(), systemPrompt, `conversation_log:\n"""`+logText+`"""`)
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

// groqModel returns the Groq chat model to use, from GROQ_MODEL, defaulting to a
// model the account is known to have access to. (The old "llama-3.3-70b-versatile"
// is not available to our key — see GET /openai/v1/models.) Must support JSON mode.
func groqModel() string {
	if m := strings.TrimSpace(os.Getenv("GROQ_MODEL")); m != "" {
		return m
	}
	return "openai/gpt-oss-20b"
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

	rawContent, err := s.callGroqChat(groqModel(), systemPrompt, `conversation_log:\n"""`+logText+`"""`)
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

