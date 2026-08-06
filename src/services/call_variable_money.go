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

// moneyVariableKeys are the flow variables that carry a baht amount and must be
// spoken as Thai baht/satang. overdue_installment is deliberately excluded — it
// is a count of installments, not money.
var moneyVariableKeys = []string{"total_debt", "total_interest", "total_fine"}

// normalizeMoneyVariables converts any money field that arrived as a bare number
// (e.g. "1000.5") into spoken Thai baht/satang, in place. Empty fields are left
// empty — nothing is filled in on the caller's behalf. Values already in spoken
// Thai form pass through unchanged (normalizeAmountToThai only converts numeric
// input), so it is safe to call unconditionally and is idempotent.
func normalizeMoneyVariables(vars map[string]string) {
	for _, k := range moneyVariableKeys {
		if v := strings.TrimSpace(vars[k]); v != "" {
			vars[k] = normalizeAmountToThai(v)
		}
	}
}

// normalizeMoneyVariablesAny is the map[string]any counterpart for the direct
// make-call path, where a money field may arrive as a JSON number (float64).
func normalizeMoneyVariablesAny(vars map[string]any) {
	for _, k := range moneyVariableKeys {
		if v := strings.TrimSpace(getStringVal(vars, k)); v != "" {
			vars[k] = normalizeAmountToThai(v)
		}
	}
}
