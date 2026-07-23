package services

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"go-fiber-template/src/client"

	fiberlog "github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type callProcessService struct {
	CallSessionsRepository  repositories.ICallSessionsRepository
	CallListItemsRepository repositories.ICallListItemsRepository
	DebtorsRepository       repositories.IDebtorsRepository
	CallRecordsRepository   repositories.ICallRecordsRepository
	CallAttemptsRepository  repositories.ICallAttemptsRepository
	OutboundClient          client.IOutboundBotnoiClient
}

type ICallProcessService interface {
	ProcessSession(sessionID string) error
	PauseSession(sessionID string) error
	StopSession(sessionID string) error
}

func NewCallProcessService(
	sessions repositories.ICallSessionsRepository,
	items repositories.ICallListItemsRepository,
	debtors repositories.IDebtorsRepository,
	records repositories.ICallRecordsRepository,
	attempts repositories.ICallAttemptsRepository,
) ICallProcessService {
	return &callProcessService{
		CallSessionsRepository:  sessions,
		CallListItemsRepository: items,
		DebtorsRepository:       debtors,
		CallRecordsRepository:   records,
		CallAttemptsRepository:  attempts,
		OutboundClient:          client.NewOutboundBotnoiClient("", "", ""),
	}
}

// thaiNumbers maps each digit to its Thai spoken word (digit-by-digit reading).
var thaiNumbers = map[rune]string{
	'0': "ศูนย์", '1': "หนึ่ง", '2': "สอง", '3': "สาม", '4': "สี่",
	'5': "ห้า", '6': "หก", '7': "เจ็ด", '8': "แปด", '9': "เก้า",
}

// toThaiDigitSpeech reads a string digit-by-digit in Thai (e.g. policy numbers).
// Non-digit separators (- _ /) are dropped; English letters are upper-cased.
func toThaiDigitSpeech(value string) string {
	normalized := ""
	for _, ch := range value {
		if ch != ' ' && ch != '\t' && ch != '\n' {
			normalized += string(ch)
		}
	}
	if normalized == "" {
		return value
	}

	hasDigit := false
	parts := []string{}
	for _, ch := range normalized {
		if word, ok := thaiNumbers[ch]; ok {
			parts = append(parts, word)
			hasDigit = true
			continue
		}
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			parts = append(parts, strings.ToUpper(string(ch)))
			continue
		}
		if ch == '-' || ch == '_' || ch == '/' {
			continue
		}
		parts = append(parts, string(ch))
	}

	if !hasDigit {
		return value
	}
	return strings.Join(parts, " ")
}

// prepareDebtorVariables clones the debtor variables, converts policy_no to
// digit-by-digit Thai speech, and adds a Thai-formatted date_today field.
func prepareDebtorVariables(input map[string]string) map[string]string {
	vars := map[string]string{}
	for k, v := range input {
		vars[k] = v
	}

	if policyNo, ok := vars["policy_no"]; ok {
		raw := strings.TrimSpace(policyNo)
		if raw != "" {
			vars["policy_no_raw"] = raw
			vars["policy_no"] = toThaiDigitSpeech(raw)
		}
	}

	vars["date_today"] = formatThaiDate(time.Now())
	return vars
}

// formatThaiDate formats a date in Thai Buddhist-era style, e.g. "วันจันทร์ ที่ 17 มิถุนายน 2569".
func formatThaiDate(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err == nil {
		t = t.In(loc)
	}
	weekdays := map[time.Weekday]string{
		time.Sunday: "วันอาทิตย์", time.Monday: "วันจันทร์", time.Tuesday: "วันอังคาร",
		time.Wednesday: "วันพุธ", time.Thursday: "วันพฤหัสบดี", time.Friday: "วันศุกร์",
		time.Saturday: "วันเสาร์",
	}
	months := []string{"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
		"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}
	buddhistYear := t.Year() + 543
	return fmt.Sprintf("%s ที่ %d %s %d", weekdays[t.Weekday()], t.Day(), months[int(t.Month())], buddhistYear)
}

const (
	staleThreshold = 5 * time.Minute
	asrProvider    = "botnoi-aws-th-noise-classifier-v17c"
)

func (sv *callProcessService) PauseSession(sessionID string) error {
	return sv.CallSessionsRepository.UpdateCallSession(sessionID, entities.CallSessionDataModel{
		Status:       "paused",
		ErrorMessage: strPtr("Paused by user"),
	})
}

func (sv *callProcessService) StopSession(sessionID string) error {
	now := time.Now().UTC()
	return sv.CallSessionsRepository.UpdateCallSession(sessionID, entities.CallSessionDataModel{
		Status:       "stopped",
		CompletedAt:  &now,
		ErrorMessage: strPtr("Stopped by user"),
	})
}

func (sv *callProcessService) ProcessSession(sessionID string) error {
	fiberlog.Infof("[Session %s] Starting processing...", sessionID)

	session, err := sv.CallSessionsRepository.FindByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	if session.Status != "running" {
		fiberlog.Infof("[Session %s] Status is %s, stopping.", sessionID, session.Status)
		return nil
	}

	if !isWithinBusinessHours(session.Settings) {
		fiberlog.Infof("[Session %s] Outside business hours, pausing.", sessionID)
		return sv.CallSessionsRepository.UpdateCallSession(sessionID, entities.CallSessionDataModel{
			Status:       "paused",
			ErrorMessage: strPtr("Paused: Outside business hours"),
		})
	}

	callingItems, err := sv.CallListItemsRepository.FindByStatus(session.WorkspaceID, session.UserID, "calling")
	if err != nil {
		return err
	}

	now := time.Now()
	var staleIDs []string
	for _, item := range *callingItems {
		if item.CalledAt == nil || item.CalledAt.IsZero() || now.Sub(*item.CalledAt) > staleThreshold {
			staleIDs = append(staleIDs, item.ID)
		}
	}

	if len(staleIDs) > 0 {
		fiberlog.Infof("[Session %s] Resetting %d stale items", sessionID, len(staleIDs))
		sv.CallListItemsRepository.UpdateManyStatus(staleIDs, "failed", "Call timed out", false)
		for _, sid := range staleIDs {
			sv.CallAttemptsRepository.UpdateStatusByListItemID(sid, "calling", "failed", "Call timed out", false, "Stale timeout (5 min)")
		}

		// Stale items never got a webhook to bump these counters themselves
		// (see webhook.go), so without this the session can never reach
		// completed_calls + failed_calls == total_calls and the frontend's
		// "Calls in Progress" banner gets stuck forever.
		session.FailedCalls += len(staleIDs)
		session.UpdatedAt = time.Now().UTC()
		if err := sv.CallSessionsRepository.UpdateCallSession(sessionID, *session); err != nil {
			fiberlog.Errorf("[Session %s] Failed to persist stale-item counter update: %s", sessionID, err)
		}
	}

	activeCallingCount := len(*callingItems) - len(staleIDs)
	maxConcurrent := session.Settings.ConcurrentCalls
	if maxConcurrent == 0 {
		maxConcurrent = 5
	}
	availableSlots := maxConcurrent - activeCallingCount
	if availableSlots < 0 {
		availableSlots = 0
	}

	fiberlog.Infof("[Session %s] calling=%d max=%d slots=%d", sessionID, activeCallingCount, maxConcurrent, availableSlots)

	if availableSlots == 0 {
		fiberlog.Infof("[Session %s] No slots, waiting for webhook.", sessionID)
		return nil
	}

	pendingItems, err := sv.CallListItemsRepository.FindPendingBySlot(session.WorkspaceID, session.UserID, availableSlots)
	if err != nil {
		return err
	}
	// Normalize to an empty list (never nil) when nothing is pending, so the
	// top-up logic below can still run and append to it.
	if pendingItems == nil {
		pendingItems = &[]entities.CallListItemModel{}
	}

	// FindPendingBySlot is capped at availableSlots but may return fewer items.
	// When AutoCall is enabled, top up any leftover slots with debtors that have
	// never been contacted yet (last_contact_at is null) by creating a pending
	// call_list_item for each and appending them to the back of pendingItems, so
	// they pass through the normal flow below.
	if session.Settings.AutoCall && len(*pendingItems) < availableSlots {
		need := availableSlots - len(*pendingItems)
		if newItems := sv.queueNeverContactedDebtors(session, need); len(newItems) > 0 {
			combined := append(*pendingItems, newItems...)
			pendingItems = &combined
		}
	}

	if len(*pendingItems) == 0 {
		waiting, _ := sv.CallListItemsRepository.CountWaitingRetry(session.WorkspaceID, session.UserID)
		if waiting > 0 {
			fiberlog.Infof("[Session %s] %d retries waiting, keep running.", sessionID, waiting)
			return nil
		}
		fiberlog.Infof("[Session %s] No pending items, completing.", sessionID)
		completedAt := time.Now().UTC()
		return sv.CallSessionsRepository.UpdateCallSession(sessionID, entities.CallSessionDataModel{
			Status:      "completed",
			CompletedAt: &completedAt,
		})
	}

	fiberlog.Infof("[Session %s] Processing %d items...", sessionID, len(*pendingItems))

	debtorIDs := make([]string, 0, len(*pendingItems))
	for _, item := range *pendingItems {
		debtorIDs = append(debtorIDs, item.DebtorID)
	}

	debtors, err := sv.DebtorsRepository.FindByIDs(debtorIDs)
	if err != nil {
		return err
	}
	debtorMap := map[string]entities.DebtorModel{}
	for _, d := range *debtors {
		debtorMap[d.ID] = d
	}

	isTestMode := session.Settings.TestMode

	itemIDs := make([]string, 0, len(*pendingItems))
	for _, item := range *pendingItems {
		itemIDs = append(itemIDs, item.ID)
	}
	sv.markManyCalling(itemIDs)

	var wg sync.WaitGroup
	var mu sync.Mutex
	completedCount, failedCount := 0, 0

	for _, item := range *pendingItems {
		wg.Add(1)
		go func(item entities.CallListItemModel) {
			defer wg.Done()
			debtor, ok := debtorMap[item.DebtorID]
			if !ok {
				sv.CallListItemsRepository.UpdateManyStatus([]string{item.ID}, "failed", "Debtor not found", false)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}
			if debtor.IsBlocked {
				sv.CallListItemsRepository.UpdateManyStatus([]string{item.ID}, "completed", "Blocked", false)
				return
			}

			success := sv.placeCall(sessionID, session, item, debtor, isTestMode)
			mu.Lock()
			if success {
				completedCount++
			} else {
				failedCount++
			}
			mu.Unlock()
		}(item)
	}
	wg.Wait()

	fiberlog.Infof("[Session %s] Batch done: %d ok, %d failed", sessionID, completedCount, failedCount)

	if isTestMode {
		sv.CallSessionsRepository.UpdateCallSession(sessionID, entities.CallSessionDataModel{
			CompletedCalls: session.CompletedCalls + completedCount,
			FailedCalls:    session.FailedCalls + failedCount,
		})
	}

	moreItems, _ := sv.CallListItemsRepository.FindPendingBySlot(session.WorkspaceID, session.UserID, maxConcurrent)
	currentCalling, _ := sv.CallListItemsRepository.FindByStatus(session.WorkspaceID, session.UserID, "calling")
	callingNow := 0
	if currentCalling != nil {
		callingNow = len(*currentCalling)
	}

	if moreItems != nil && len(*moreItems) > 0 && (maxConcurrent-callingNow) > 0 {
		fiberlog.Infof("[Session %s] More items + slots, continuing...", sessionID)
		return sv.ProcessSession(sessionID)
	}

	if callingNow == 0 {
		waiting, _ := sv.CallListItemsRepository.CountWaitingRetry(session.WorkspaceID, session.UserID)
		if waiting > 0 {
			return nil
		}
		fiberlog.Infof("[Session %s] All processed, completing.", sessionID)
		completedAt := time.Now().UTC()
		return sv.CallSessionsRepository.UpdateCallSession(sessionID, entities.CallSessionDataModel{
			Status:      "completed",
			CompletedAt: &completedAt,
		})
	}

	return nil
}

// queueNeverContactedDebtors finds up to `need` debtors that have never been
// contacted (last_contact_at is null) and don't already have a pending
// call_list_item, then creates a pending call_list_item for each — exactly as the
// frontend does when a user queues a debtor — so they can pass through the normal
// calling flow. It returns the newly created items.
func (sv *callProcessService) queueNeverContactedDebtors(session *entities.CallSessionDataModel, need int) []entities.CallListItemModel {
	if need <= 0 {
		return nil
	}

	// Exclude debtors that already have a pending call_list_item so we never
	// queue the same debtor twice.
	var excludeIDs []string
	if pending, err := sv.CallListItemsRepository.FindByStatus(session.WorkspaceID, session.UserID, "pending"); err == nil && pending != nil {
		for _, it := range *pending {
			excludeIDs = append(excludeIDs, it.DebtorID)
		}
	}

	debtors, err := sv.DebtorsRepository.FindNeverContacted(session.WorkspaceID, session.UserID, excludeIDs, need)
	if err != nil || debtors == nil || len(*debtors) == 0 {
		return nil
	}

	now := time.Now().UTC()
	newItems := make([]entities.CallListItemModel, 0, len(*debtors))
	for _, d := range *debtors {
		item := entities.CallListItemModel{
			ID:           uuid.NewString(),
			UserID:       session.UserID,
			DebtorID:     d.ID,
			WorkspaceID:  session.WorkspaceID,
			Status:       "pending",
			RetryCount:   0,
			ScheduledAt:  &now,
			DebtorPhone:  d.PhoneNumber,
			DebtorName:   DebtorDisplayName(&d),
			DebtorAmount: DebtorDisplayAmount(&d),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := sv.CallListItemsRepository.Insert(item); err != nil {
			fiberlog.Errorf("[Session %s] queueNeverContacted: insert item for debtor %s: %s", session.ID, d.ID, err)
			continue
		}
		newItems = append(newItems, item)
	}

	if len(newItems) > 0 {
		fiberlog.Infof("[Session %s] Queued %d never-contacted debtors to fill slots", session.ID, len(newItems))
	}
	return newItems
}

func (sv *callProcessService) markManyCalling(ids []string) {
	for _, id := range ids {
		nowTime := time.Now().UTC()
		sv.CallListItemsRepository.Update(id, entities.CallListItemModel{
			Status:   "calling",
			CalledAt: &nowTime,
		})
	}
}

func (sv *callProcessService) placeCall(
	sessionID string,
	session *entities.CallSessionDataModel,
	item entities.CallListItemModel,
	debtor entities.DebtorModel,
	isTestMode bool,
) bool {
	if isTestMode {
		// Simulate processing time 1-3s
		time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)

		// Random outcome: 40% confirmed, 20% declined, 20% no_response, 10% no_answer, 10% failed
		r := rand.Float64()
		var mockStatus entities.CallStatus
		var mockOutcome string
		pickedUp := false

		switch {
		case r < 0.4:
			mockStatus = entities.StatusConfirmed
			mockOutcome = "ยืนยันชำระ"
			pickedUp = true
		case r < 0.6:
			mockStatus = entities.StatusDeclined
			mockOutcome = "ปฏิเสธ"
			pickedUp = true
		case r < 0.8:
			mockStatus = entities.StatusNoResponse
			mockOutcome = "ไม่ตอบ"
			pickedUp = true
		case r < 0.9:
			mockStatus = entities.StatusNoAnswer
			mockOutcome = "ไม่รับสาย"
			pickedUp = false
		default:
			mockStatus = entities.StatusFailed
			mockOutcome = "โทรไม่สำเร็จ"
			pickedUp = false
		}

		// Create mock call record
		sv.CallRecordsRepository.InsertCallRecord(entities.CallRecordDataModel{
			PhoneNumber:  debtor.PhoneNumber,
			BotnoiCallID: fmt.Sprintf("test_%d", time.Now().UnixNano()),
			Status:       mockStatus,
			UserID:       session.UserID,
			WorkspaceID:  session.WorkspaceID,
		})

		// Update call list item
		sv.CallListItemsRepository.UpdateManyStatus([]string{item.ID}, string(mockStatus), mockOutcome, pickedUp)

		// Update debtor stats (start from existing values, then +1 by condition)
		nowTime := time.Now().UTC()
		stats := entities.DebtorStatsUpdate{
			ContactAttempts:    debtor.ContactAttempts + 1,
			SuccessfulContacts: debtor.SuccessfulContacts,
			PickedUpCount:      debtor.PickedUpCount,
			NotPickedUpCount:   debtor.NotPickedUpCount,
			LastContactAt:      &nowTime,
			LastResponse:       mockOutcome,
			CallOutcome:        string(mockStatus),
			CallAnswered:       boolPtr(pickedUp),
		}
		if pickedUp {
			stats.PickedUpCount = debtor.PickedUpCount + 1
			stats.SuccessfulContacts = debtor.SuccessfulContacts + 1
		} else {
			stats.NotPickedUpCount = debtor.NotPickedUpCount + 1
		}
		sv.DebtorsRepository.UpdateStats(item.DebtorID, stats)

		return mockStatus != entities.StatusFailed
	}

	vars := prepareDebtorVariables(debtor.Variables)
	// Fill missing variables from debtor columns / mock defaults so incomplete
	// debtor data never produces a broken flow.
	vars = applyDefaultCallVariables(vars, &debtor)
	vars["bot_type"] = "{{in_init_conversation}}"
	vars["intent"] = "{{in_init_conversation}}"


	// outbound_id is what comes back to us in the webhook, so we use it as the call id.
	outboundID := "outbound_" + item.ID

	payload := entities.OutboundBotnoiDataModel{
		OutboundID: outboundID,
		Flow: buildFlow(outboundID,
			vars["name"],
			vars["car_detail"],
			vars["province"],
			vars["total_debt"],
			vars["total_interest"],
			vars["total_fine"],
			vars["overdue_installment"],
			vars["bot_type"],
			vars["intent"]),
		PhoneNumber: "3525" + debtor.PhoneNumber,
		BotID:       os.Getenv("BOT_ID"),
		BotType:     os.Getenv("BOT_TYPE"),
		Intent:      "in_init_conversation",
		// The partner /outbound contract only requires outbound_id, phonenumber,
		// flow, bot_id. The extra call-config fields below are kept (commented)
		// for future use — re-enable if the partner API stops applying defaults.
		 EventID:          "event_" + sessionID + "_" + item.ID,
		 SourcePhone:      "3525" + debtor.PhoneNumber,
		 Speaker:          "212",
		 Language:         "th",
		 AgentPhoneNumber: "0800000000",
		 Speed:            "1",
		 TTS:              "voicebot-premium",
		 ASRProvider:      "botnoi-th-noise-classifier-C",
		 ASRLanguageCode:  "th",
		 ASRTimeout:       5,
		 FalseTimeoutSec:  "1",
		 FalseSilenceSec:  "0.1",
		 TrueSilenceSec:   "0.25",
		 Interruptible:    boolToStr(session.Settings.Interruptible),
	}

	if err := sv.OutboundClient.MakeCall(payload); err != nil {
		sv.CallListItemsRepository.UpdateManyStatus([]string{item.ID}, "failed", err.Error(), false)
		return false
	}

	// Generate the call record id ourselves so we can link it back to the item/attempt.
	callRecordID := uuid.NewString()

	sv.CallRecordsRepository.InsertCallRecord(entities.CallRecordDataModel{
		ID:           callRecordID,
		PhoneNumber:  debtor.PhoneNumber,
		BotnoiCallID: outboundID,
		Status:       entities.StatusPending,
		UserID:       session.UserID,
		WorkspaceID:  session.WorkspaceID,
	})

	// Link call_record_id back onto the list item (keep status "calling" until webhook).
	calledTime := time.Now().UTC()
	sv.CallListItemsRepository.Update(item.ID, entities.CallListItemModel{
		Status:       "calling",
		CallOutcome:  "Call initiated - awaiting response",
		CalledAt:     &calledTime,
		CallRecordID: callRecordID,
	})

	// attempt_number = current retry_count + 1
	attemptNumber := item.RetryCount + 1

	sv.CallAttemptsRepository.Insert(entities.CallAttemptModel{
		CallListItemID: item.ID,
		CallRecordID:   callRecordID,
		UserID:         session.UserID,
		WorkspaceID:    session.WorkspaceID,
		AttemptNumber:  attemptNumber,
		Status:         "calling",
		CallOutcome:    "Call initiated - awaiting response",
		PickedUp:       boolPtr(false),
	})

	// Update debtor contact attempt count (real result comes later via webhook).
	nowTimeUTC := time.Now().UTC()
	sv.DebtorsRepository.UpdateStats(item.DebtorID, entities.DebtorStatsUpdate{
		ContactAttempts:    debtor.ContactAttempts + 1,
		SuccessfulContacts: debtor.SuccessfulContacts,
		PickedUpCount:      debtor.PickedUpCount,
		NotPickedUpCount:   debtor.NotPickedUpCount,
		LastContactAt:      &nowTimeUTC,
		LastResponse:       debtor.LastResponse,
		CallOutcome:        debtor.CallOutcome,
		CallAnswered:       debtor.CallAnswered,
	})

	return true
}

func isWithinBusinessHours(settings entities.CallSessionSettings) bool {
	if settings.TestMode {
		return true
	}
	if !settings.BusinessHoursOnly {
		return true
	}

	now := time.Now().UTC()
	local := now.Add(time.Duration(settings.TimezoneOffset) * time.Minute)
	currentDay := int(local.Weekday())

	inDay := false
	for _, d := range settings.BusinessDays {
		if d == currentDay {
			inDay = true
			break
		}
	}
	if !inDay {
		return false
	}

	current := local.Hour()*60 + local.Minute()
	start := parseHHMM(settings.BusinessHoursStart, 9*60)
	end := parseHHMM(settings.BusinessHoursEnd, 18*60)
	return current >= start && current <= end
}

func parseHHMM(s string, def int) int {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return def
	}
	return h*60 + m
}

func strPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }

func boolToStr(b bool) string {
	if b {
		return "True"
	}
	return "False"
}
