package services

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"go-fiber-template/domain/entities"
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
// (e.g. "1000.5", or a JSON number the caller sent) into spoken Thai baht/satang.
// It runs on the FINAL value regardless of where it came from — request payload,
// debtor column, or mock default — because a number reaches Botnoi as literal
// "point five" otherwise. Values already in spoken Thai form pass through
// unchanged (normalizeAmountToThai only converts numeric input), so it is safe to
// call unconditionally and is idempotent.
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

// defaultCallVariables are the base/mock values (mirroring cmd/seed) used when a
// debtor is missing a variable the call flow needs. Every key consumed by
// buildFlow must have an entry here so a call can always be placed. Edit this
// map to change what the bot speaks for missing data.
// Keys map to the bot's parameters:
//
//	name (customer_name) = ชื่อคน, car_detail = ทะเบียนรถ, province = จังหวัด,
//	total_debt = จำนวนเงินที่ค้างชำระ, total_interest = ดอกเบี้ย,
//	total_fine = เบี้ยที่ปรับ, overdue_installment = งวดค้างชำระ
var defaultCallVariables = map[string]string{
	"name":                "คุณสมชาย",
	"car_detail":          "กก1111",
	"province":            "กรุงเทพมหานคร",
	"total_debt":          "3000",
	"total_interest":      "300",
	"total_fine":          "500",
	"overdue_installment": "2",
}

// applyDefaultCallVariables fills any missing/empty flow variable so a call can
// still be placed when debtor data is incomplete. Resolution order per key:
// value already in vars → real debtor data (name/total_debt columns, amount
// aliases) → defaultCallVariables.
func applyDefaultCallVariables(vars map[string]string, debtor *entities.DebtorModel) map[string]string {
	if vars == nil {
		vars = map[string]string{}
	}

	// Real debtor data beats mock defaults.
	if strings.TrimSpace(vars["name"]) == "" && debtor != nil {
		fullName := strings.TrimSpace(strings.TrimSpace(debtor.Name) + " " + strings.TrimSpace(debtor.LastName))
		if fullName != "" {
			vars["name"] = fullName
		}
	}
	if strings.TrimSpace(vars["total_debt"]) == "" {
		// Mirror DebtorDisplayAmount: accept the frontend's amount aliases first.
		// These are numeric strings, so speak them as Thai baht/satang rather than
		// sending a bare "1000.5" that Botnoi would read as "point five".
		for _, key := range []string{"amount", "outstanding_amount"} {
			if raw := strings.TrimSpace(vars[key]); raw != "" {
				vars["total_debt"] = normalizeAmountToThai(raw)
				break
			}
		}
		if strings.TrimSpace(vars["total_debt"]) == "" && debtor != nil && debtor.TotalDebt > 0 {
			vars["total_debt"] = formatThaiBahtSatang(debtor.TotalDebt)
		}
	}

	for key, def := range defaultCallVariables {
		if strings.TrimSpace(vars[key]) == "" {
			vars[key] = def
		}
	}

	// Speak every money field as Thai baht/satang, whatever its source.
	normalizeMoneyVariables(vars)
	return vars
}

// applyDefaultVoicebotVariables is the map[string]any counterpart used by the
// direct make-call API, where there is no debtor record to fall back on.
func applyDefaultVoicebotVariables(variables map[string]any) map[string]any {
	if variables == nil {
		variables = map[string]any{}
	}
	if strings.TrimSpace(getStringVal(variables, "total_debt")) == "" {
		for _, key := range []string{"amount", "outstanding_amount"} {
			if raw := strings.TrimSpace(getStringVal(variables, key)); raw != "" {
				variables["total_debt"] = normalizeAmountToThai(raw)
				break
			}
		}
	}
	for key, def := range defaultCallVariables {
		if strings.TrimSpace(getStringVal(variables, key)) == "" {
			variables[key] = def
		}
	}

	// Speak every money field as Thai baht/satang, whatever its source.
	normalizeMoneyVariablesAny(variables)
	return variables
}
