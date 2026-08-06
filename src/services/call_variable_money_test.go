package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestNormalizeMoneyVariables_SpeaksNumbersLeavesRest(t *testing.T) {
	vars := map[string]string{
		"total_debt":          "1000.5",              // bare number -> spoken
		"total_interest":      "1,234.50",            // commas stripped -> spoken
		"total_fine":          "500 บาท",             // already spoken -> untouched
		"overdue_installment": "2",                   // not money -> untouched
		"name":                "คุณลูกค้า",           // not money -> untouched
	}
	normalizeMoneyVariables(vars)

	assert.Equal(t, "1000 บาท 50 สตางค์", vars["total_debt"])
	assert.Equal(t, "1234 บาท 50 สตางค์", vars["total_interest"])
	assert.Equal(t, "500 บาท", vars["total_fine"])
	assert.Equal(t, "2", vars["overdue_installment"])
	assert.Equal(t, "คุณลูกค้า", vars["name"])
}

// An empty money field is left empty — no default is filled in.
func TestNormalizeMoneyVariables_EmptyStaysEmpty(t *testing.T) {
	vars := map[string]string{"total_debt": ""}
	normalizeMoneyVariables(vars)
	assert.Equal(t, "", vars["total_debt"])
	// A field that was never set stays absent.
	_, ok := vars["total_interest"]
	assert.False(t, ok)
}

// Direct make-call path: a JSON number (float64) for a money field is spoken.
func TestNormalizeMoneyVariablesAny_JSONNumber(t *testing.T) {
	vars := map[string]any{"total_debt": 1000.5}
	normalizeMoneyVariablesAny(vars)
	assert.Equal(t, "1000 บาท 50 สตางค์", vars["total_debt"])
}
