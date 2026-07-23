package services

import (
	"strconv"
	"strings"

	"go-fiber-template/domain/entities"
)

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
		for _, key := range []string{"amount", "outstanding_amount"} {
			if raw := strings.TrimSpace(vars[key]); raw != "" {
				vars["total_debt"] = strings.ReplaceAll(raw, ",", "")
				break
			}
		}
		if strings.TrimSpace(vars["total_debt"]) == "" && debtor != nil && debtor.TotalDebt > 0 {
			vars["total_debt"] = strconv.FormatFloat(debtor.TotalDebt, 'f', -1, 64)
		}
	}

	for key, def := range defaultCallVariables {
		if strings.TrimSpace(vars[key]) == "" {
			vars[key] = def
		}
	}
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
				variables["total_debt"] = strings.ReplaceAll(raw, ",", "")
				break
			}
		}
	}
	for key, def := range defaultCallVariables {
		if strings.TrimSpace(getStringVal(variables, key)) == "" {
			variables[key] = def
		}
	}
	return variables
}
