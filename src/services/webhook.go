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
)

type ClassifyResult struct {
	StatusID   int    `json:"status_id"`
	StatusName string `json:"status_name"`
	Category   string `json:"category"`
	Reason     string `json:"reason"`
}

var CONVERSATION_CATEGORIES = []struct {
	ID    int
	Name  string
	Thai  string
	Group string
}{
	{1, "Planned More Than 3", "วางแผนชำระเกิน 3 วัน", "main"},
	{2, "Promised to Pay", "รับปากชำระ", "main"},
	{3, "Restructure Requested", "ขอปรับโครงสร้างหนี้", "main"},
	{4, "Inconvenient (With Date)", "ไม่สะดวก (มีนัดหมาย)", "main"},
	{16, "Inconvenient (Without Date)", "ไม่สะดวก (ไม่มีนัดหมาย)", "main"},
	{5, "Already Paid", "ชำระเรียบร้อยแล้ว", "main"},
	{6, "Not Reached", "ติดต่อไม่ได้", "main"},
	{7, "Refused", "ปฏิเสธ", "main"},
	{8, "Not Convenient", "ไม่สะดวกคุย", "sub"},
	{9, "Wrong Person", "ไม่ใช่ผู้เอาประกัน", "sub"},
	{10, "Call Later", "นัดหมายให้ติดต่อใหม่", "sub"},
	{11, "Transfer", "ขอคุยกับเจ้าหน้าที่", "sub"},
	{12, "Background Noise", "เสียงแทรก/เสียงรบกวน", "sub"},
	{13, "Silence", "ลูกค้าเงียบ", "sub"},
	{14, "Dropped Call", "สายหลุดระหว่างสนทนา", "sub"},
	{15, "Out of Topic", "พูดเรื่องอื่น", "sub"},
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
	if callID == "" {
		callID = payload.CallID
	}
	status := payload.Status
	action := payload.Action
	conversationLog := payload.ConversationLog
	audioURL := payload.AudioURL
	phoneNumber := payload.PhoneNumber

	// Extract phone number from audio_url if missing (format: ..._PHONE.wav)
	if phoneNumber == "" && audioURL != "" {
		re := regexp.MustCompile(`_(\d+)\.wav$`)
		match := re.FindStringSubmatch(audioURL)
		if len(match) > 1 {
			phoneNumber = match[1]
		}
	}

	if callID == "" && phoneNumber == "" {
		log.Warn("Webhook received with no identifiable data")
		return nil
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

	pickedUp := hasUserSpoken || isSilence || s.contains([]string{string(entities.StatusConfirmed), string(entities.StatusDeclined), string(entities.StatusNoResponse), string(entities.StatusCompleted)}, string(mappedStatus))

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

	// --- AI Categorization ---
	aiResult := s.classifyCall(payload, conversationLog)
	aiCategory := aiResult.Category

	// Resolve Owner (UserID, WorkspaceID)
	var resolvedUserID, resolvedWorkspaceID string

	// 1. Try from CallRecord
	var callRecord *entities.CallRecordDataModel
	if callID != "" {
		records, _ := s.CallRecordsService.GetAllCallRecords(entities.CallRecordFilter{BotnoiCallID: callID})
		if records != nil && len(*records) > 0 {
			callRecord = &(*records)[0]
			resolvedUserID = callRecord.UserID
			resolvedWorkspaceID = callRecord.WorkspaceID
			if phoneNumber == "" {
				phoneNumber = callRecord.PhoneNumber
			}
		}
	}

	// 2. Fallback: Try from Debtor (WorkspaceID resolve)
	if resolvedWorkspaceID == "" && phoneNumber != "" {
		debtor, _ := s.DebtorService.GetDebtorByPhoneNumber(phoneNumber)
		if debtor != nil {
			resolvedUserID = debtor.UserID
			resolvedWorkspaceID = debtor.WorkspaceID
		}
	}

	// Update Call Record and related entities
	if callRecord != nil {
		var duration int
		if payload.CallDuration != nil {
			duration = s.toInt(payload.CallDuration)
		} else if payload.Duration != nil {
			duration = s.toInt(payload.Duration)
		}

		callRecord.Status = mappedStatus
		var resultData interface{} = payload
		callRecord.ResultData = &resultData
		callRecord.CallDuration = duration
		callRecord.AppointmentDate = payload.AppointmentDate
		callRecord.AppointmentTime = payload.AppointmentTime
		callRecord.UpdatedAt = time.Now().UTC()

		s.CallRecordsService.UpdateCallRecord(callRecord.ID, *callRecord)

		// Update Call List Items
		items, _ := s.CallListItemService.GetCallListItemsByWorkspace(callRecord.WorkspaceID)
		if items != nil {
			for _, item := range *items {
				if item.CallRecordID == callRecord.ID {
					item.Status = finalStatus
					item.CallOutcome = callOutcome
					item.PickedUp = &pickedUp
					item.AICategory = aiCategory
					item.NextRetryAt = nil
					notesObj := map[string]string{
						"audio_url":        audioURL,
						"conversation_log": conversationLog,
					}
					notesJSON, _ := json.Marshal(notesObj)
					item.Notes = string(notesJSON)
					item.UpdatedAt = time.Now().UTC()
					s.CallListItemService.UpdateCallListItem(item.ID, item)

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
							ConversationLog: conversationLog,
							AudioURL:        audioURL,
							CallDuration:    duration,
						}
						s.CallAttemptService.CreateAttempt(newAttempt)
					}
				}
			}
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
						debtor.AcceptCount++
						debtor.LastResponse = "accept"
					} else if mappedStatus == entities.StatusDeclined {
						debtor.RejectCount++
						debtor.LastResponse = "reject"
					} else if mappedStatus == entities.StatusNoResponse || (mappedStatus == entities.StatusCompleted && pickedUp) {
						debtor.OtherCount++
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
					s.CallSessionService.UpdateCallSession(session.ID, session)

					// Trigger the next batch via the call-process service directly.
					// Run in a goroutine so the webhook response is not blocked by
					// the (potentially long-running, recursive) session processing.
					sessionID := session.ID
					go func() {
						_ = s.CallProcessService.ProcessSession(sessionID)
					}()
				}
			}
		}
	}

	return nil
}

func (s *webhookService) classifyCall(payload entities.WebhookPayload, logText string) ClassifyResult {
	apiKey := os.Getenv("LOVABLE_API_KEY")
	if apiKey == "" || logText == "" || len(logText) < 5 {
		return ClassifyResult{Category: "Not Reached", Reason: "No log or API key missing"}
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
		return ClassifyResult{Category: cat, Reason: "System status: " + rawStatus}
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
			return ClassifyResult{Category: "Silence", Reason: "Customer picked up but remained silent (ASR TIMEOUT)"}
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
			return ClassifyResult{Category: "Background Noise", Reason: "Detected audio-quality keywords"}
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

Main Outcomes (business result of the call — ALWAYS PREFER THESE):
- Planned More Than 3     → Customer indicates a payment plan / will pay, but the agreed date is MORE THAN 3 days from the call date (e.g. "อาทิตย์หน้า", "เกิน 3 วัน", "เดือนหน้า", "อีก 5 วัน", or any specific date > 3 days away). Also use for vague acknowledgement of the debt with no clear near-term commitment.
- Promised to Pay        → Customer explicitly confirms payment WITHIN 3 days from the call date (today, tomorrow, "พรุ่งนี้", "มะรืน", "อีก 2 วัน", or any specific date ≤ 3 days away).
- Restructure Requested  → Customer asks for debt restructuring, installment plans, payment negotiation, partial payment, deferral, or settlement discussion.
- Inconvenient (With Date)    → Customer says it is not convenient right now BUT provides a specific callback date/time (e.g. "call me back tomorrow at 3pm", "next Monday morning"). A concrete schedule is agreed.
- Inconvenient (Without Date) → Customer says it is not convenient and does NOT provide any specific callback date/time (vague "call me later", "not now", "I'm busy").
- Already Paid           → Customer states the payment has already been completed/settled.
- Not Reached            → Customer could not actually be contacted (no answer, line dead, voicemail, unreachable, hung up before any meaningful exchange).
- Refused                → Customer clearly refuses to pay, denies the debt outright, or terminates the conversation in clear refusal.

Conversation Behaviors (use ONLY when no clear business outcome above exists):
- Not Convenient   → Customer says it is not a convenient time but did not commit to a callback time.
- Wrong Person     → Customer says this is not the policyholder / wrong number.
- Call Later       → Customer vaguely asks to be called another time without a fixed schedule.
- Transfer         → Customer asks to speak to a human agent / staff.
- Background Noise → Audio quality issues, loud background, customer cannot hear clearly, transcript dominated by noise.
- Silence          → Customer remained silent throughout / no verbal response.
- Dropped Call     → Call disconnected with no meaningful customer interaction (pure cutoff).
- Out of Topic     → Customer kept talking about unrelated topics with no resolution.

CRITICAL CLASSIFICATION RULES
1. MANDATORY DEBT DISCLOSURE CHECK (HIGHEST PRIORITY): Verify Bot has DISCLOSED "Debt Details".
2. HANDLING "NOT CONVENIENT" & CALLBACK REQUESTS: Use With Date/Without Date labels.
3. PAYMENT CLASSIFICATION (ONLY AFTER DISCLOSURE): ≤ 3 days = Promised to Pay, > 3 days = Planned More Than 3.

Output format (STRICT JSON):
{
  "status_name": "<exact English label from the list>",
  "confidence": <number between 0 and 1>,
  "reason": "<short explanation>"
}`

	type aiRequest struct {
		Model          string                 `json:"model"`
		Messages       []map[string]string    `json:"messages"`
		ResponseFormat map[string]interface{} `json:"response_format"`
		Temperature    float64                `json:"temperature"`
	}

	reqBody := aiRequest{
		Model: "google/gemini-2.5-flash",
		Messages: []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": `conversation_log:\n"""` + logText + `"""`},
		},
		ResponseFormat: map[string]interface{}{"type": "json_object"},
		Temperature:    0,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post("https://ai.gateway.lovable.dev/v1/chat/completions", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Errorf("AI Classify Error: Request failed: %v", err)
		return ClassifyResult{Category: "Not Reached", Reason: "AI request failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("AI Classify Error: Status %d, Body: %s", resp.StatusCode, string(body))
		return ClassifyResult{Category: "Not Reached", Reason: "AI response error"}
	}

	body, _ := io.ReadAll(resp.Body)
	var aiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(body, &aiResponse)

	if len(aiResponse.Choices) > 0 {
		var content struct {
			StatusName string `json:"status_name"`
			Reason     string `json:"reason"`
		}
		json.Unmarshal([]byte(aiResponse.Choices[0].Message.Content), &content)

		aiName := strings.ToLower(strings.TrimSpace(content.StatusName))
		if aiName == "acknowledged" || aiName == "acknowledge" {
			aiName = "planned more than 3"
		}

		for _, c := range CONVERSATION_CATEGORIES {
			if strings.ToLower(c.Name) == aiName {
				return ClassifyResult{
					StatusID:   c.ID,
					StatusName: c.Name,
					Category:   c.Name,
					Reason:     content.Reason,
				}
			}
		}
	}

	return ClassifyResult{Category: "Not Reached", Reason: "Defaulted or unmatched"}
}

func (s *webhookService) extractCallbackDate(logText string) string {
	apiKey := os.Getenv("LOVABLE_API_KEY")
	if apiKey == "" || logText == "" || len(logText) < 5 {
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

	type aiRequest struct {
		Model          string                 `json:"model"`
		Messages       []map[string]string    `json:"messages"`
		ResponseFormat map[string]interface{} `json:"response_format"`
		Temperature    float64                `json:"temperature"`
	}

	reqBody := aiRequest{
		Model: "google/gemini-3-flash-preview",
		Messages: []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": `conversation_log:\n"""` + logText + `"""`},
		},
		ResponseFormat: map[string]interface{}{"type": "json_object"},
		Temperature:    0,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post("https://ai.gateway.lovable.dev/v1/chat/completions", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Errorf("AI Date Extract Error: Status %d, Body: %s", resp.StatusCode, string(body))
		}
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var aiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(body, &aiResponse)

	if len(aiResponse.Choices) > 0 {
		var content struct {
			DateCon string `json:"date_con"`
		}
		json.Unmarshal([]byte(aiResponse.Choices[0].Message.Content), &content)
		return content.DateCon
	}

	return ""
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
