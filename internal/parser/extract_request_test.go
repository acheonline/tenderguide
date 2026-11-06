package parser

import "testing"

func TestExtractRequestProcedureNumber(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"Изучив Извещение о проведении Конкурса, опубликованное в ЕИС за № 32616122324 от «17» июня 2026 г.",
	}}}
	got := extractRequest(in)
	if got.ProcedureNumber == nil || *got.ProcedureNumber != "32616122324" {
		t.Fatalf("procedure_number = %v", got.ProcedureNumber)
	}
	if got.ProcedurePublishedAt == nil || *got.ProcedurePublishedAt != "2026-06-17" {
		t.Fatalf("procedure_published_at = %v", got.ProcedurePublishedAt)
	}
}

func TestExtractRequestLetterNumberNotProcedureNumber(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"№197 От «01» июля 2026 года",
	}}}
	got := extractRequest(in)
	if got.ProcedureNumber != nil {
		t.Fatalf("procedure_number should be nil, got %v", *got.ProcedureNumber)
	}
	if got.LetterNumber == nil || *got.LetterNumber != "197" {
		t.Fatalf("letter_number = %v", got.LetterNumber)
	}
	if got.LetterDate == nil || *got.LetterDate != "2026-07-01" {
		t.Fatalf("letter_date = %v", got.LetterDate)
	}
}

func TestExtractRequestProcedureName(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"предлагает заключить договор на оказание услуг по обеспечению безопасности веб-приложений АО «ТЭК СПб», (наименование предмета Конкурса)",
	}}}
	got := extractRequest(in)
	if got.ProcedureName == nil {
		t.Fatal("procedure_name = nil")
	}
	if *got.ProcedureName != "оказание услуг по обеспечению безопасности веб-приложений АО «ТЭК СПб»" {
		t.Fatalf("procedure_name = %q", *got.ProcedureName)
	}
}

func TestExtractRequestProcedureNameTypo(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"предлагает заключить договор на на оказание услуг, (наименование)",
	}}}
	got := extractRequest(in)
	if got.ProcedureName == nil {
		t.Fatal("procedure_name = nil")
	}
	if *got.ProcedureName != "оказание услуг" {
		t.Fatalf("procedure_name = %q", *got.ProcedureName)
	}
}

func TestExtractRequestPrices(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"на условиях на общую сумму: 14 414 600,00 (Четырнадцать миллионов) рублей 00 копеек (общая сумма заявки цифрами и прописью, рублей, с учетом НДС)",
		"2 013 000,00 (Два миллиона тринадцать тысяч) рублей 00 копеек (сумма НДС цифрами и прописью, рублей)",
		"12 401 600,00 (Двенадцать миллионов) рублей 00 копеек. (общая сумма заявки цифрами и прописью, рублей, без учета НДС)",
	}}}
	got := extractRequest(in)
	if got.PriceWithVAT == nil || *got.PriceWithVAT != "14414600.00" {
		t.Fatalf("price_with_vat = %v", got.PriceWithVAT)
	}
	if got.VAT == nil || *got.VAT != "2013000.00" {
		t.Fatalf("vat = %v", got.VAT)
	}
	if got.PriceWithoutVAT == nil || *got.PriceWithoutVAT != "12401600.00" {
		t.Fatalf("price_without_vat = %v", got.PriceWithoutVAT)
	}
}
