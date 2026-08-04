package services

import (
	"testing"

	"go-fiber-template/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestApplyDefaultCallVariables_NilVarsNilDebtor(t *testing.T) {
	vars := applyDefaultCallVariables(nil, nil)

	// Every flow key must be present so buildFlow never gets an empty slot.
	assert.Equal(t, "คุณสมชาย", vars["name"])
	assert.Equal(t, "กก1111", vars["car_detail"])
	assert.Equal(t, "กรุงเทพมหานคร", vars["province"])
	assert.Equal(t, "3000", vars["total_debt"])
	assert.Equal(t, "300", vars["total_interest"])
	assert.Equal(t, "500", vars["total_fine"])
	assert.Equal(t, "2", vars["overdue_installment"])
}

func TestApplyDefaultCallVariables_DebtorColumnsBeatDefaults(t *testing.T) {
	debtor := &entities.DebtorModel{
		Name:      "สมชาย",
		LastName:  "ใจดี",
		TotalDebt: 2500.75,
	}
	vars := applyDefaultCallVariables(map[string]string{}, debtor)

	assert.Equal(t, "สมชาย ใจดี", vars["name"])
	// Derived from the numeric column -> spoken Thai baht/satang, not "2500.75".
	assert.Equal(t, "2500 บาท 75 สตางค์", vars["total_debt"])
	// Columns with no counterpart still fall to defaults.
	assert.Equal(t, "กก1111", vars["car_detail"])
	assert.Equal(t, "กรุงเทพมหานคร", vars["province"])
}

func TestApplyDefaultCallVariables_ExistingVarsUntouched(t *testing.T) {
	debtor := &entities.DebtorModel{Name: "ผิด", TotalDebt: 999}
	vars := applyDefaultCallVariables(map[string]string{
		"name":       "คุณลูกค้าจริง",
		"total_debt": "3000",
		"car_detail": "Toyota Vios",
		"province":   "เชียงใหม่",
	}, debtor)

	assert.Equal(t, "คุณลูกค้าจริง", vars["name"])
	assert.Equal(t, "3000", vars["total_debt"])
	assert.Equal(t, "Toyota Vios", vars["car_detail"])
	assert.Equal(t, "เชียงใหม่", vars["province"])
}

func TestApplyDefaultCallVariables_AmountAliases(t *testing.T) {
	// outstanding_amount (as seeded by cmd/seed) fills total_debt before mocks,
	// converted to spoken Thai baht/satang.
	vars := applyDefaultCallVariables(map[string]string{
		"outstanding_amount": "1,234.50",
	}, nil)
	assert.Equal(t, "1234 บาท 50 สตางค์", vars["total_debt"])
}

func TestFormatThaiBahtSatang(t *testing.T) {
	cases := []struct {
		amount float64
		want   string
	}{
		{1000.5, "1000 บาท 50 สตางค์"},
		{3000, "3000 บาท"},
		{2500.75, "2500 บาท 75 สตางค์"},
		{0.999, "0 บาท 99 สตางค์"}, // truncated, not rounded up to 1 บาท
		{1000.05, "1000 บาท 5 สตางค์"},
		{0.29, "0 บาท 29 สตางค์"}, // float 0.28999… must not lose a satang
		{0, "0 บาท"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, formatThaiBahtSatang(tc.amount), "amount=%v", tc.amount)
	}
}

// A spoken-Thai amount already in the request must pass through untouched — the
// backend must never re-derive it from a numeric column.
func TestApplyDefaultCallVariables_SpokenAmountPassesThrough(t *testing.T) {
	debtor := &entities.DebtorModel{TotalDebt: 1000.5}
	vars := applyDefaultCallVariables(map[string]string{
		"total_debt": "1000 บาท 50 สตางค์",
	}, debtor)
	assert.Equal(t, "1000 บาท 50 สตางค์", vars["total_debt"])
}

func TestApplyDefaultVoicebotVariables(t *testing.T) {
	vars := applyDefaultVoicebotVariables(map[string]any{
		"name":   "สมหญิง",
		"amount": "4,000",
	})

	assert.Equal(t, "สมหญิง", vars["name"])
	// "amount" alias is numeric -> spoken Thai (whole baht, no satang).
	assert.Equal(t, "4000 บาท", vars["total_debt"])
	assert.Equal(t, "กก1111", vars["car_detail"])
	assert.Equal(t, "กรุงเทพมหานคร", vars["province"])
	assert.Equal(t, "300", vars["total_interest"])
}
