package parser

import "testing"

func TestExtractBankAccounts(t *testing.T) {
	in := []Input{
		{
			Tables: [][]Row{
				{
					Row{Cells: []string{"13.", "Банковские реквизиты", "Филиал Центральный Банка ВТБ (ПАО), расчетный счет 40702810200810114144, корреспондентский счет 30101810145250000411, БИК банка 044525411"}},
				},
			},
		},
	}
	got := extractBankAccounts(in)
	if len(got) != 1 {
		t.Fatalf("want 1 bank_account, got %d", len(got))
	}
	b := got[0]
	if b.BIK == nil || *b.BIK != "044525411" {
		t.Fatalf("bik = %v", b.BIK)
	}
	if b.Account == nil || *b.Account != "40702810200810114144" {
		t.Fatalf("account = %v", b.Account)
	}
	if b.CorrespondentAccount == nil || *b.CorrespondentAccount != "30101810145250000411" {
		t.Fatalf("corr = %v", b.CorrespondentAccount)
	}
}
