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

// The bot script carries no unit words, so the backend supplies them: amounts as
// baht/satang, interest/fine with their label, installments with "งวด".
func TestApplyFlowVariables(t *testing.T) {
	vars := map[string]string{
		"total_debt":          "1000.5",
		"total_interest":      "100.3",
		"total_fine":          "500",
		"overdue_installment": "2",
		"name":                "คุณลูกค้า", // untouched
	}
	applyFlowVariables(vars)

	assert.Equal(t, "1000 บาท 50 สตางค์", vars["total_debt"])
	assert.Equal(t, "ดอกเบี้ย 100 บาท 30 สตางค์", vars["total_interest"])
	assert.Equal(t, "เบี้ยปรับ 500 บาท", vars["total_fine"])
	assert.Equal(t, "2งวด", vars["overdue_installment"])
	assert.Equal(t, "คุณลูกค้า", vars["name"])
}

// interest/fine with no real value are dropped entirely, so the bot never reads a
// dangling "ดอกเบี้ย" / "เบี้ยปรับ".
func TestApplyFlowVariables_BlankInterestFineOmitted(t *testing.T) {
	for _, blank := range []string{"", "0", "-", "0.00", "  "} {
		vars := map[string]string{"total_interest": blank, "total_fine": blank}
		applyFlowVariables(vars)
		assert.Equal(t, "", vars["total_interest"], "interest=%q", blank)
		assert.Equal(t, "", vars["total_fine"], "fine=%q", blank)
	}
}

// An empty installment stays empty — no lone "งวด".
func TestApplyFlowVariables_EmptyInstallmentStaysEmpty(t *testing.T) {
	vars := map[string]string{"overdue_installment": ""}
	applyFlowVariables(vars)
	assert.Equal(t, "", vars["overdue_installment"])
}

// Direct make-call path: JSON numbers (float64) are formatted, and a zero fine is
// dropped.
func TestApplyFlowVariablesAny_JSONNumbers(t *testing.T) {
	vars := map[string]any{
		"total_debt":     1000.5,
		"total_interest": 100.3,
		"total_fine":     0, // zero -> omitted
	}
	applyFlowVariablesAny(vars)

	assert.Equal(t, "1000 บาท 50 สตางค์", vars["total_debt"])
	assert.Equal(t, "ดอกเบี้ย 100 บาท 30 สตางค์", vars["total_interest"])
	assert.Equal(t, "", vars["total_fine"])
}

func TestIsBlankAmount(t *testing.T) {
	blank := []string{"", "  ", "-", "0", "0.00", "0.0"}
	for _, s := range blank {
		assert.True(t, isBlankAmount(s), "should be blank: %q", s)
	}
	real := []string{"100.3", "0.01", "1,000", "-5"}
	for _, s := range real {
		assert.False(t, isBlankAmount(s), "should be real: %q", s)
	}
}
