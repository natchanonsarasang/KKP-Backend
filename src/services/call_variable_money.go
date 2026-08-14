package services

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// formatThaiBahtSatang renders a numeric amount as spoken Thai baht/satang, e.g.
// 1000.5 -> "1000 บาท 50 สตางค์" and 3000 -> "3000 บาท". Satang is TRUNCATED, not
// rounded: 0.999 -> "0 บาท 99 สตางค์" (never rounds a satang up into a baht). The
// tiny epsilon absorbs binary float error so an amount stored as e.g. 0.28999…
// still yields 29 สตางค์.
func formatThaiBahtSatang(amount float64) string {
	if amount < 0 {
		amount = 0
	}
	totalSatang := int64(math.Floor(amount*100 + 1e-6))
	baht := totalSatang / 100
	satang := totalSatang % 100
	if satang == 0 {
		return fmt.Sprintf("%d บาท", baht)
	}
	return fmt.Sprintf("%d บาท %d สตางค์", baht, satang)
}

// normalizeAmountToThai converts a raw amount string into spoken Thai baht/satang
// when it is numeric (e.g. "1,000.5" -> "1000 บาท 50 สตางค์"). A non-numeric value
// is assumed to already be in spoken form and is returned unchanged (minus commas).
func normalizeAmountToThai(raw string) string {
	cleaned := strings.ReplaceAll(strings.TrimSpace(raw), ",", "")
	if cleaned == "" {
		return cleaned
	}
	if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
		return formatThaiBahtSatang(f)
	}
	return cleaned
}

// isBlankAmount reports whether a money field carries no meaningful value and
// should be dropped entirely — empty, "-", or a value that is 0. This lets the
// caller omit the whole clause (label included) instead of the bot reading a
// pointless "ดอกเบี้ย 0 บาท".
func isBlankAmount(raw string) bool {
	s := strings.TrimSpace(raw)
	if s == "" || s == "-" {
		return true
	}
	if f, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64); err == nil {
		return f == 0
	}
	return false
}

// labeledAmount returns "<label> <spoken amount>" (e.g. "ดอกเบี้ย 100 บาท 30 สตางค์"),
// or "" when the amount is blank/zero so the bot skips the label entirely.
func labeledAmount(label, raw string) string {
	if isBlankAmount(raw) {
		return ""
	}
	return label + " " + normalizeAmountToThai(raw)
}

// applyFlowVariables formats the amount/count flow variables to the exact spoken
// text the bot expects, IN PLACE. The bot script no longer carries the unit words,
// so the backend now supplies them:
//   - total_debt          -> Thai baht/satang (the script keeps its own label)
//   - total_interest      -> "ดอกเบี้ย <amount>", omitted entirely when zero/blank
//   - total_fine          -> "เบี้ยปรับ <amount>", omitted entirely when zero/blank
//   - overdue_installment -> "<n>งวด", empty stays empty
//
// Any field left empty is sent empty so the bot reads nothing there.
func applyFlowVariables(vars map[string]string) {
	if v := strings.TrimSpace(vars["total_debt"]); v != "" {
		vars["total_debt"] = normalizeAmountToThai(v)
	}
	vars["total_interest"] = labeledAmount("ดอกเบี้ย", vars["total_interest"])
	vars["total_fine"] = labeledAmount("เบี้ยปรับ", vars["total_fine"])
	if v := strings.TrimSpace(vars["overdue_installment"]); v != "" {
		vars["overdue_installment"] = v + "งวด"
	}
}

// applyFlowVariablesAny is the map[string]any counterpart for the direct make-call
// path, where a value may arrive as a JSON number (float64).
func applyFlowVariablesAny(vars map[string]any) {
	if v := strings.TrimSpace(getStringVal(vars, "total_debt")); v != "" {
		vars["total_debt"] = normalizeAmountToThai(v)
	}
	vars["total_interest"] = labeledAmount("ดอกเบี้ย", getStringVal(vars, "total_interest"))
	vars["total_fine"] = labeledAmount("เบี้ยปรับ", getStringVal(vars, "total_fine"))
	if v := strings.TrimSpace(getStringVal(vars, "overdue_installment")); v != "" {
		vars["overdue_installment"] = v + "งวด"
	}
}
