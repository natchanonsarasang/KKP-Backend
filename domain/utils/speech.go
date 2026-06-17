package utils

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

var THAI_DIGIT_WORDS = map[rune]string{
	'0': "ศูนย์",
	'1': "หนึ่ง",
	'2': "สอง",
	'3': "สาม",
	'4': "สี่",
	'5': "ห้า",
	'6': "หก",
	'7': "เจ็ด",
	'8': "แปด",
	'9': "เก้า",
}

var thaiMonths = []string{
	"มกราคม", "กุมภาพันธ์", "มีนาคม",
	"เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน",
	"ตุลาคม", "พฤศจิกายน", "ธันวาคม",
}

var thaiWeekdays = []string{
	"วันอาทิตย์", "วันจันทร์", "วันอังคาร",
	"วันพุธ", "วันพฤหัสบดี", "วันศุกร์",
	"วันเสาร์",
}

func ToThaiDigitSpeech(value string) string {
	normalized := strings.TrimSpace(strings.Join((strings.Fields(value)), ""))
	if normalized == "" {
		return value
	}

	var parts []string
	hasDigit := false
	for _, ch := range normalized {
		if word, ok := THAI_DIGIT_WORDS[ch]; ok {
			parts = append(parts, word)
			hasDigit = true
			continue
		}

		if unicode.IsLetter(ch) && ch <= unicode.MaxASCII {
			parts = append(parts, string(ch))
		}

		if ch == ' ' || ch == '-' || ch == '_' {
			continue
		}

		parts = append(parts, string(ch))
	}

	if hasDigit {
		return strings.Join(parts, " ")
	}

	return value
}

func ThaiTodayString() string {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	return fmt.Sprintf(
		"%s ที่ %d %s %d",
		thaiWeekdays[now.Weekday()],
		now.Day(),
		thaiMonths[now.Month()-1],
		now.Year()+543,
	)
}
