package parser

import "testing"

func TestExtractParticipantInnKpp(t *testing.T) {
	in := []Input{{Paragraphs: []string{"000 «Центр…» ИНН/КПП: 7453299127/ 781401001"}}}
	got := extractParticipant(in)
	if got == nil {
		t.Fatal("participant = nil")
	}
	if got.INN == nil || *got.INN != "7453299127" {
		t.Fatalf("inn = %v", got.INN)
	}
	if got.KPP == nil || *got.KPP != "781401001" {
		t.Fatalf("kpp = %v", got.KPP)
	}
}

func TestExtractParticipantFullName(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"Общество с ограниченной ответственностью «Центр информационных технологий и информационной безопасности», (полное наименование)",
	}}}
	got := extractParticipant(in)
	if got == nil || got.FullName == nil {
		t.Fatal("full_name = nil")
	}
	if *got.FullName != "Общество с ограниченной ответственностью «Центр информационных технологий и информационной безопасности»" {
		t.Fatalf("full_name = %q", *got.FullName)
	}
}

func TestExtractParticipantOgrn(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"ИНН/КПП: 7453299127/ 781401001",
		"ОГРН: 1167456110744",
	}}}
	got := extractParticipant(in)
	if got == nil || got.OGRN == nil || *got.OGRN != "1167456110744" {
		t.Fatalf("ogrn = %v", got)
	}
}

func TestExtractParticipantContact(t *testing.T) {
	in := []Input{{Paragraphs: []string{
		"ИНН/КПП: 7453299127/ 781401001",
		"info@citib.ru, http://www.citib.ru",
	}}}
	got := extractParticipant(in)
	if got == nil {
		t.Fatal("participant = nil")
	}
	if got.Email == nil || *got.Email != "info@citib.ru" {
		t.Fatalf("email = %v", got.Email)
	}
	if got.Website == nil || *got.Website != "http://www.citib.ru" {
		t.Fatalf("website = %v", got.Website)
	}
}
