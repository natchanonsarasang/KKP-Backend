package services

import (
	"math"
	"strconv"
	"strings"
	"time"

	"go-fiber-template/domain/entities"

	fiberlog "github.com/gofiber/fiber/v2/log"
)

// The Botnoi webhook that used to close the loop for a call is not finished yet,
// so this file polls the batch status endpoint (GET /outbound/batch/{batch_id})
// to learn each call's outcome and then finalizes it — the same responsibilities
// the webhook has (update records, advance the session, trigger the next call).
//
// This is intentionally self-contained so it can be deleted wholesale once the
// webhook is live: remove the `go sv.pollBatchUntilDone(...)` call in placeCall
// and this file.

const (
	// How often to re-check the batch status while waiting for a result.
	batchPollInterval = 5 * time.Second
	// Give up polling after this long and finalize the call as failed. Kept at the
	// stale threshold so a call is never left "calling" forever if a result never
	// lands (the stale sweep in ProcessSession is the restart-safe backstop).
	batchPollTimeout = staleThreshold
)

// pollBatchUntilDone polls the batch for this call until it reports a terminal
// outcome (or the timeout elapses), then finalizes it. Runs in its own goroutine.
func (sv *callProcessService) pollBatchUntilDone(
	sessionID string,
	item entities.CallListItemModel,
	debtor entities.DebtorModel,
	callRecordID string,
	batchID string,
) {
	tag := "[BatchPoll " + batchID + "]"
	deadline := time.Now().Add(batchPollTimeout)

	for {
		if time.Now().After(deadline) {
			fiberlog.Warnf("%s timed out after %s, finalizing as failed", tag, batchPollTimeout)
			sv.finalizeCallFromBatch(sessionID, item, debtor, callRecordID, batchID, nil)
			return
		}

		time.Sleep(batchPollInterval)

		status, err := sv.OutboundClient.GetBatchStatus(batchID)
		if err != nil {
			fiberlog.Warnf("%s status fetch failed, will retry: %s", tag, err)
			continue
		}

		call := findBatchCall(status, debtor.PhoneNumber)
		if call == nil || !isTerminalCallStatus(call.Status) {
			continue
		}

		fiberlog.Infof("%s call terminal: status=%q had_answered=%t duration=%q", tag, call.Status, call.HadAnswered, call.CallDuration)
		sv.finalizeCallFromBatch(sessionID, item, debtor, callRecordID, batchID, call)
		return
	}
}

// findBatchCall returns the call matching the dialed phone number, falling back to
// the first call in the batch (our batches only ever hold one number).
func findBatchCall(status *entities.OutboundBatchStatusResponse, phone string) *entities.OutboundBatchCall {
	if status == nil || len(status.Calls) == 0 {
		return nil
	}
	for i := range status.Calls {
		if status.Calls[i].PhoneNumber == phone {
			return &status.Calls[i]
		}
	}
	return &status.Calls[0]
}

// isTerminalCallStatus reports whether a call-level status is a final outcome
// (anything that is not an in-progress/ringing/queued state).
func isTerminalCallStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "pending", "queued", "dialing", "ringing", "calling", "initiated", "in_progress", "in-progress", "connecting":
		return false
	default:
		return true
	}
}

// classifyBatchCall maps a call-level batch result onto our CallStatus taxonomy.
// There is no conversation transcript here, so this is call-level only (no AI
// category) — that resumes when the webhook is available.
func classifyBatchCall(call *entities.OutboundBatchCall) (mapped entities.CallStatus, pickedUp bool) {
	if call == nil {
		return entities.StatusFailed, false
	}

	status := strings.ToLower(strings.TrimSpace(call.Status))
	callState := strings.ToUpper(strings.TrimSpace(call.CallStateValue))
	pickedUp = call.HadAnswered

	switch {
	case status == "hanged_up" || status == "hangup" || status == "hung_up":
		// The debtor answered and then hung up — counts as a pickup, but the call
		// itself is a failed outcome (see finalizeCallFromBatch, mirroring webhook).
		return entities.StatusHangedUp, true
	case call.AMDVoicemailDetected || status == "voicemail":
		return entities.StatusVoicemail, false
	case status == "rejected":
		return entities.StatusRejected, false
	case status == "busy" || callState == "BUSY":
		return entities.StatusBusy, false
	case status == "no_answer" || status == "no answer" || status == "noanswer" || status == "timeout":
		return entities.StatusNoAnswer, false
	case status == "failed" || status == "error":
		return entities.StatusFailed, false
	case call.HadAnswered:
		return entities.StatusCompleted, true
	default:
		return entities.StatusNoAnswer, false
	}
}

var batchOutcomeMap = map[entities.CallStatus]string{
	entities.StatusConfirmed:  "Confirmed",
	entities.StatusDeclined:   "Declined",
	entities.StatusNoResponse: "No Response",
	entities.StatusNoAnswer:   "No Answer",
	entities.StatusCompleted:  "Completed",
	entities.StatusFailed:     "Failed",
	entities.StatusBusy:       "Busy",
	entities.StatusRejected:   "Rejected",
	entities.StatusVoicemail:  "Voicemail",
	entities.StatusHangedUp:   "Hangup",
}

// finalizeCallFromBatch closes the loop for one call using its polled batch
// result (or nil on timeout): it updates the call_record, call_list_item and
// call_attempt, refreshes the debtor's pickup stats, drops the queue row, advances
// the session counters and triggers the next call. Mirrors webhook.ProcessWebhook
// minus the conversation/AI parts.
func (sv *callProcessService) finalizeCallFromBatch(
	sessionID string,
	item entities.CallListItemModel,
	debtor entities.DebtorModel,
	callRecordID string,
	batchID string,
	call *entities.OutboundBatchCall,
) {
	tag := "[BatchFinalize " + batchID + "]"

	mappedStatus, pickedUp := classifyBatchCall(call)
	outcome := batchOutcomeMap[mappedStatus]
	if outcome == "" {
		outcome = "Unknown"
	}
	// A hangup is a pickup for the debtor's stats but a failed call outcome, so it
	// stays "failed" even though pickedUp is true (mirrors webhook.go).
	finalStatus := "failed"
	if pickedUp && mappedStatus != entities.StatusHangedUp {
		finalStatus = "success"
	}
	duration := 0
	if call != nil {
		duration = parseDurationSeconds(call.CallDuration)
	}

	// 1. call_record
	if rec, err := sv.CallRecordsRepository.FindByID(callRecordID); err == nil && rec != nil {
		rec.Status = mappedStatus
		rec.CallDuration = duration
		if call != nil {
			var resultData interface{} = call
			rec.ResultData = &resultData
		}
		rec.UpdatedAt = time.Now().UTC()
		if err := sv.CallRecordsRepository.UpdateCallRecord(rec.ID, *rec); err != nil {
			fiberlog.Errorf("%s update call_record %s failed: %s", tag, rec.ID, err)
		}
	}

	// 2. call_list_item + call_attempt
	sv.CallListItemsRepository.UpdateManyStatus([]string{item.ID}, finalStatus, outcome, pickedUp)
	sv.CallAttemptsRepository.UpdateStatusByListItemID(item.ID, "calling", finalStatus, outcome, pickedUp, "")

	// 3. debtor pickup stats. ContactAttempts was already incremented at dial time
	// (see placeCall), so here we only record the pickup outcome the dial couldn't
	// yet know. Refetch to avoid clobbering concurrent updates.
	current := &debtor
	if fresh, err := sv.DebtorsRepository.FindByID(item.DebtorID); err == nil && fresh != nil {
		current = fresh
	}
	nowTime := time.Now().UTC()
	stats := entities.DebtorStatsUpdate{
		ContactAttempts:    current.ContactAttempts,
		SuccessfulContacts: current.SuccessfulContacts,
		PickedUpCount:      current.PickedUpCount,
		NotPickedUpCount:   current.NotPickedUpCount,
		LastContactAt:      &nowTime,
		LastResponse:       current.LastResponse,
		CallOutcome:        string(mappedStatus),
		CallAnswered:       boolPtr(pickedUp),
	}
	if pickedUp {
		stats.PickedUpCount = current.PickedUpCount + 1
		stats.SuccessfulContacts = current.SuccessfulContacts + 1
	} else {
		stats.NotPickedUpCount = current.NotPickedUpCount + 1
	}
	sv.DebtorsRepository.UpdateStats(item.DebtorID, stats)

	// 4. drop the queue row so the next call's KKP_Data fetch reads the next debtor.
	if err := sv.CallQueueService.DeleteByOutboundID(batchID); err != nil {
		fiberlog.Errorf("%s delete queue row %q failed: %s", tag, batchID, err)
	}

	// 5. advance the session counters (fetch fresh so we don't wipe fields).
	session, err := sv.CallSessionsRepository.FindByID(sessionID)
	if err != nil || session == nil {
		fiberlog.Warnf("%s session %s not found, skipping advance", tag, sessionID)
		return
	}
	if session.Status == "running" {
		if finalStatus == "success" {
			session.CompletedCalls++
			if mappedStatus == entities.StatusConfirmed {
				session.ConfirmedCalls++
			}
		} else {
			session.FailedCalls++
		}
		session.UpdatedAt = time.Now().UTC()
		if err := sv.CallSessionsRepository.UpdateCallSession(session.ID, *session); err != nil {
			fiberlog.Errorf("%s update session %s failed: %s", tag, session.ID, err)
		}

		fiberlog.Infof("%s finalized: status=%s outcome=%q pickedUp=%t → triggering next call", tag, finalStatus, outcome, pickedUp)
		// Trigger the next call. Runs in its own goroutine so a long recursive
		// ProcessSession never blocks this finalizer.
		go func() {
			if err := sv.ProcessSession(sessionID); err != nil {
				fiberlog.Errorf("%s ProcessSession %s failed: %s", tag, sessionID, err)
			}
		}()
	}
}

// parseDurationSeconds parses Botnoi's call_duration (seconds as a string, e.g.
// "11.699") into whole seconds. Returns 0 when it can't be parsed.
func parseDurationSeconds(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return int(math.Round(f))
	}
	return 0
}
