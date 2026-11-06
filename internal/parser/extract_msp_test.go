package parser

import "testing"

func TestExtractMSPInfoAbsent(t *testing.T) {
	in := []Input{{Paragraphs: []string{"какой-то текст"}}}
	if got := extractMSPInfo(in); got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func TestExtractMSPInfoPresent(t *testing.T) {
	in := []Input{
		{Paragraphs: []string{"Декларация о соответствии Участника закупки критериям отнесения к субъектам малого и среднего предпринимательства"}},
		{
			Tables: [][]Row{
				{
					Row{Cells: []string{"7.", "Среднесписочная численность работников", "до 100", "от 101 до 250", "1"}},
					Row{Cells: []string{"8.", "Доход за предшествующий календарный год", "800", "2000", "67,898"}},
				},
			},
		},
	}
	got := extractMSPInfo(in)
	if got == nil {
		t.Fatal("msp = nil")
	}
	if got.IsSMP == nil || !*got.IsSMP {
		t.Fatalf("is_smp = %v", got.IsSMP)
	}
	if got.AverageHeadcount == nil || *got.AverageHeadcount != 1 {
		t.Fatalf("headcount = %v", got.AverageHeadcount)
	}
	if got.RevenuePreviousYear == nil || *got.RevenuePreviousYear != 67898000 {
		t.Fatalf("revenue = %v", got.RevenuePreviousYear)
	}
	if got.SMPCategory == nil || *got.SMPCategory != "micro" {
		t.Fatalf("category = %v", got.SMPCategory)
	}
}

func TestCalcSMPCategory(t *testing.T) {
	h1, r1 := int64(1), int64(67898000)
	h2, r2 := int64(85), int64(500000000)
	h3, r3 := int64(200), int64(900000000)
	cases := []struct {
		h, r *int64
		want string
	}{
		{&h1, &r1, "micro"},
		{&h2, &r2, "small"},
		{&h3, &r3, "medium"},
	}
	for _, c := range cases {
		got := calcSMPCategory(c.h, c.r)
		if got == nil || *got != c.want {
			t.Errorf("calcSMPCategory(%d, %d) = %v, want %q", *c.h, *c.r, got, c.want)
		}
	}
}
