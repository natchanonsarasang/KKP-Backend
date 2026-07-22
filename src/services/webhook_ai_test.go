package services

import (
	"os"
	"testing"

	"go-fiber-template/domain/entities"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// These tests exercise the AI transcript-analysis functions in isolation.
// classifyCall/extractCallbackDate only read GROQ_API_KEY and make an HTTP call
// to Groq — they never touch the repositories/MongoDB — so we can call them on a
// zero-value webhookService (all deps nil). Nothing here writes to the database.
//
// They make a REAL request to Groq, so they skip when GROQ_API_KEY is absent.
// Run with:  go test ./src/services/ -run TestAI -v

func newAITestService(t *testing.T) *webhookService {
	// Load the project .env (test cwd is src/services) so GROQ_API_KEY is available.
	_ = godotenv.Load("../../.env")
	if os.Getenv("GROQ_API_KEY") == "" {
		t.Skip("GROQ_API_KEY not set — skipping live Groq test")
	}
	return &webhookService{} // nil repos: these methods don't touch the DB
}

// isKnownCategory reports whether name is one of the fixed CONVERSATION_CATEGORIES
// labels. A valid label means Groq responded and the JSON parsed end-to-end.
func isKnownCategory(name string) bool {
	for _, c := range CONVERSATION_CATEGORIES {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestAIClassifyCall(t *testing.T) {
	s := newAITestService(t)

	cases := []struct {
		name string
		log  string
		want string // expected category for clear-cut cases
	}{
		{
			name: "willing to pay",
			log:  "Bot: สวัสดีค่ะ ติดต่อเรื่องยอดค้างชำระ 4000 บาทค่ะ User: พรุ่งนี้ผมโอนให้เลยครับ",
			want: "Convenient to Pay",
		},
		{
			name: "refuses to pay",
			log:  "Bot: แจ้งยอดค้างชำระค่ะ User: ไม่จ่าย ไม่มีเงิน อย่าโทรมาอีก วางสายนะ",
			want: "Not Convenient to Pay",
		},
		{
			name: "asks to restructure",
			log:  "Bot: ยอดค้างชำระ 4000 บาทค่ะ User: ขอผ่อนเป็นงวดได้ไหมครับ จ่ายทีเดียวไม่ไหว",
			want: "Not Convenient to Pay",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := s.classifyCall(entities.WebhookPayload{Status: "completed"}, tc.log)

			// Log what the AI actually said so quality is eyeball-able even when it passes.
			t.Logf("category=%q reason=%q", res.Category, res.Reason)

			// Hard assertion: the call succeeded and produced a valid taxonomy label.
			assert.True(t, isKnownCategory(res.Category),
				"expected a known category, got %q (AI call likely failed)", res.Category)

			// Soft check: flag but don't fail if the label differs — LLM output can vary.
			if res.Category != tc.want {
				t.Logf("NOTE: expected %q but got %q", tc.want, res.Category)
			}
		})
	}
}

func TestAIExtractCallbackDate(t *testing.T) {
	s := newAITestService(t)

	// "พรุ่งนี้" (tomorrow) must yield a concrete YYYY-MM-DD date.
	log := "Bot: สะดวกชำระเมื่อไหร่คะ User: พรุ่งนี้ผมจ่ายครับ"
	date := s.extractCallbackDate(log)

	t.Logf("extracted date_con=%q", date)
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, date,
		"expected an ISO date for 'พรุ่งนี้', got %q", date)
}
