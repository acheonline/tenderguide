package parser

import "testing"

func TestExtractEmpty(t *testing.T) {
	got := Extract("req-empty", nil)
	if got == nil || got.Request == nil {
		t.Fatal("nil result")
	}
	if got.Request.RequestID != "req-empty" {
		t.Fatalf("request_id = %q", got.Request.RequestID)
	}
	if got.Request.CreatedAt == "" {
		t.Fatal("created_at is empty")
	}
	if got.Participant != nil || got.Customer != nil {
		t.Fatal("participant/customer should be nil")
	}
}

func TestExtractGarbage(t *testing.T) {
	in := []Input{{Paragraphs: []string{"¡¢", "ÿþýü"}}}
	got := Extract("req-garbage", in)
	if got == nil || got.Request == nil {
		t.Fatal("nil result")
	}
	if got.Request.ProcedureNumber != nil {
		t.Fatal("procedure_number should be nil")
	}
}

func TestExtractFull(t *testing.T) {
	in := []Input{
		{Paragraphs: []string{
			"Изучив Извещение о проведении Конкурса, опубликованное в ЕИС за № [](https://zakupki.gov.ru/epz/order/notice/notice223/common-info.html?regNumber=32616122324) от «17» июня 2026 г.",
			"№197",
			"От «01» июля 2026 года",
			"предлагает заключить договор на",
			"оказание услуг по обеспечению безопасности веб-приложений АО «ТЭК СПб»,",
			"000 «Центр…» ИНН/КПП: 7453299127/ 781401001",
		}},
	}
	got := Extract("req-001", in)
	if got.Request.ProcedureNumber == nil || *got.Request.ProcedureNumber != "32616122324" {
		t.Fatalf("procedure_number = %v", got.Request.ProcedureNumber)
	}
	if got.Request.ProcedureName == nil {
		t.Fatal("procedure_name = nil")
	}
	if got.Participant == nil || got.Participant.INN == nil || *got.Participant.INN != "7453299127" {
		t.Fatalf("participant.inn = %v", got.Participant)
	}
	if got.Customer == nil || got.Customer.Name == nil {
		t.Fatal("customer should be detected")
	}
}
