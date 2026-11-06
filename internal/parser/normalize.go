package parser

import (
	"regexp"
	"strings"
)

// Мусорный префикс до первой буквы, цифры или №.
var reGarbagePrefix = regexp.MustCompile(`^[^\p{L}\d№]+`)

// reMarkdownLink вырезает []() из текста, оставляя только подпись.
var reMarkdownLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

// stripGarbage убирает мусорный префикс, неразрывные пробелы и markdown-ссылки.
func stripGarbage(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = reMarkdownLink.ReplaceAllString(s, "$1")
	return reGarbagePrefix.ReplaceAllString(s, "")
}

var monthNames = map[string]string{
	"января":   "01",
	"февраля":  "02",
	"марта":    "03",
	"апреля":   "04",
	"мая":      "05",
	"июня":     "06",
	"июля":     "07",
	"августа":  "08",
	"сентября": "09",
	"октября":  "10",
	"ноября":   "11",
	"декабря":  "12",
}

var reDateQuote = regexp.MustCompile(`«(\d{1,2})»\s+([а-я]+)\s+(\d{4})`)

func normalizeDate(s string) *string {
	m := reDateQuote.FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	return dateFromParts(m[1], m[2], m[3])
}

func dateFromParts(day, month, year string) *string {
	if len(day) == 1 {
		day = "0" + day
	}
	m, ok := monthNames[strings.ToLower(month)]
	if !ok {
		return nil
	}
	res := year + "-" + m + "-" + day
	return &res
}

var reDecimal = regexp.MustCompile(`\d[\d\s]*,\d{2}`)

func normalizeDecimal(s string) *string {
	m := reDecimal.FindString(s)
	if m == "" {
		return nil
	}
	m = strings.ReplaceAll(m, " ", "")
	m = strings.ReplaceAll(m, "\u00a0", "")
	m = strings.ReplaceAll(m, ",", ".")
	return &m
}

var rePhoneDigits = regexp.MustCompile(`\d`)

func normalizePhone(s string) *string {
	digits := rePhoneDigits.FindAllString(s, -1)
	if len(digits) < 10 {
		return nil
	}
	last10 := strings.Join(digits[len(digits)-10:], "")
	res := "+7" + last10
	return &res
}

var reDigits = regexp.MustCompile(`\d`)

func onlyDigits(s string) string {
	return strings.Join(reDigits.FindAllString(s, -1), "")
}

func normalizeINN(s string) *string {
	d := onlyDigits(s)
	if len(d) != 10 {
		return nil
	}
	return &d
}

func normalizeKPP(s string) *string {
	d := onlyDigits(s)
	if len(d) != 9 {
		return nil
	}
	return &d
}

func normalizeOGRN(s string) *string {
	d := onlyDigits(s)
	if len(d) != 13 && len(d) != 15 {
		return nil
	}
	return &d
}

var reDotDate = regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{4})`)

func normalizeDotDate(s string) *string {
	m := reDotDate.FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	res := m[3] + "-" + m[2] + "-" + m[1]
	return &res
}
