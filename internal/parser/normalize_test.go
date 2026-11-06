package parser

import "testing"

func TestStripGarbage(t *testing.T) {
	cases := map[string]string{
		"¡¢ИНН: 745": "ИНН: 745",
		"12345":        "12345",
		"текст":        "текст",
	}
	for in, want := range cases {
		if got := stripGarbage(in); got != want {
			t.Errorf("stripGarbage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeDate(t *testing.T) {
	cases := map[string]string{
		"«01» июля 2026":    "2026-07-01",
		"«31» декабря 2026": "2026-12-31",
		"«17» июня 2026":    "2026-06-17",
	}
	for in, want := range cases {
		got := normalizeDate(in)
		if got == nil || *got != want {
			t.Errorf("normalizeDate(%q) = %v, want %q", in, got, want)
		}
	}
	if v := normalizeDate("не дата"); v != nil {
		t.Errorf("normalizeDate(не дата) = %v, want nil", *v)
	}
}

func TestNormalizeDecimal(t *testing.T) {
	v := normalizeDecimal("14 414 600,00 рублей")
	if v == nil || *v != "14414600.00" {
		t.Fatalf("got %v", v)
	}
}

func TestNormalizePhone(t *testing.T) {
	cases := map[string]string{
		"+7 (351) 326 72 10": "+73513267210",
		"+7 (922)-128-69-62": "+79221286962",
	}
	for in, want := range cases {
		got := normalizePhone(in)
		if got == nil || *got != want {
			t.Errorf("normalizePhone(%q) = %v, want %q", in, got, want)
		}
	}
}
