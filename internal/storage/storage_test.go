package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/acheonline/tenderguide/internal/model"
)

func strp(s string) *string { return &s }
func intp(i int64) *int64   { return &i }
func boolp(b bool) *bool    { return &b }

func newStore(t *testing.T) *Store {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=foreign_keys(1)"
	s, err := Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpenIdempotent(t *testing.T) {
	s := newStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestNoNotNullOnBusinessFields(t *testing.T) {
	s := newStore(t)
	rows, err := s.db.Query(`PRAGMA table_info(participants)`)
	if err != nil {
		t.Fatalf("pragma: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if name == "id" {
			continue
		}
		if notnull != 0 {
			t.Errorf("participants.%s: notnull=%d, want 0", name, notnull)
		}
	}
}

func TestSaveFullAndRoundTrip(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	in := &model.All{
		Request: &model.Request{
			RequestID:            "req-001",
			ProcedureNumber:      strp("32616122324"),
			ProcedureName:        strp("оказание услуг"),
			ProcedurePublishedAt: strp("2026-06-17"),
			LetterNumber:         strp("197"),
			LetterDate:           strp("2026-07-01"),
			ValidUntil:           strp("2026-12-31"),
			PriceWithVAT:         strp("14414600.00"),
			VAT:                  strp("2013000.00"),
			PriceWithoutVAT:      strp("12401600.00"),
			PriceCurrency:        strp("RUB"),
		},
		Customer: &model.Customer{INN: nil, Name: strp("АО «ТЭК СПб»")},
		Participant: &model.Participant{
			INN:          strp("7453299127"),
			KPP:          strp("781401001"),
			OGRN:         strp("1167456110744"),
			FullName:     strp("ООО «ЦИТИБ»"),
			LegalAddress: strp("197342, СПб"),
			HasBranches:  boolp(false),
		},
		BankAccounts: []*model.BankAccount{
			{BankName: strp("ВТБ"), Account: strp("40702810200810114144"), BIK: strp("044525411")},
		},
		MSPInfo: &model.MSPInfo{
			IsSMP:               boolp(true),
			AverageHeadcount:    intp(1),
			RevenuePreviousYear: intp(67898000),
			SMPCategory:         strp("micro"),
		},
		StaffSummary: &model.StaffSummary{
			TotalStaff:       intp(85),
			ManagersCount:    intp(21),
			SpecialistsCount: intp(64),
			WorkersCount:     intp(0),
		},
		StaffMembers: []*model.StaffMember{
			{Category: strp("managers"), FullName: strp("Дубинович И.В."), Position: strp("Зам. ген. дир."), ExperienceYears: strp("3")},
			{Category: strp("specialists"), FullName: strp("Кудрявцев Д.А."), Position: strp("Руководитель направления"), ExperienceYears: strp("6")},
		},
		Beneficiaries: []*model.Beneficiary{
			{FullName: strp("Никитин Алексей Владимирович"), INN: strp("655903142602"), ShareGT10Pct: boolp(true)},
		},
	}

	pk, err := s.Save(ctx, in)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if pk == 0 {
		t.Fatal("Save returned pk=0")
	}

	out, err := s.GetByRequestID(ctx, "req-001")
	if err != nil {
		t.Fatalf("GetByRequestID: %v", err)
	}
	if out.Request.ProcedureNumber == nil || *out.Request.ProcedureNumber != "32616122324" {
		t.Errorf("procedure_number mismatch: %v", out.Request.ProcedureNumber)
	}
	if out.Request.PriceWithVAT == nil || *out.Request.PriceWithVAT != "14414600.00" {
		t.Errorf("price_with_vat mismatch: %v", out.Request.PriceWithVAT)
	}
	if out.Participant == nil || out.Participant.INN == nil || *out.Participant.INN != "7453299127" {
		t.Errorf("participant.inn mismatch")
	}
	if len(out.BankAccounts) != 1 {
		t.Errorf("bank_accounts: want 1, got %d", len(out.BankAccounts))
	}
	if out.MSPInfo == nil || out.MSPInfo.IsSMP == nil || !*out.MSPInfo.IsSMP {
		t.Errorf("msp_info.is_smp mismatch")
	}
	if out.StaffSummary == nil || out.StaffSummary.TotalStaff == nil || *out.StaffSummary.TotalStaff != 85 {
		t.Errorf("staff_summary.total_staff mismatch")
	}
	if len(out.StaffMembers) != 2 {
		t.Errorf("staff_members: want 2, got %d", len(out.StaffMembers))
	}
	if len(out.Beneficiaries) != 1 {
		t.Errorf("beneficiaries: want 1, got %d", len(out.Beneficiaries))
	}
}

func TestSaveMinimal(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	in := &model.All{Request: &model.Request{RequestID: "req-min"}}
	if _, err := s.Save(ctx, in); err != nil {
		t.Fatalf("Save minimal: %v", err)
	}

	out, err := s.GetByRequestID(ctx, "req-min")
	if err != nil {
		t.Fatalf("GetByRequestID: %v", err)
	}
	if out.Request.CustomerID != nil {
		t.Errorf("customer_id: want nil, got %v", out.Request.CustomerID)
	}
	if out.Request.ParticipantID != nil {
		t.Errorf("participant_id: want nil, got %v", out.Request.ParticipantID)
	}
	if out.Customer != nil || out.Participant != nil {
		t.Errorf("customer/participant: want nil")
	}
}

func TestSaveDuplicateRequestID(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	in := &model.All{Request: &model.Request{RequestID: "req-dup"}}
	if _, err := s.Save(ctx, in); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	_, err := s.Save(ctx, in)
	if !errors.Is(err, ErrDuplicateRequest) {
		t.Fatalf("want ErrDuplicateRequest, got %v", err)
	}
}

func TestSaveParticipantDedup(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	p := &model.Participant{INN: strp("7453299127"), FullName: strp("X")}

	for _, id := range []string{"req-a", "req-b"} {
		_, err := s.Save(ctx, &model.All{
			Request:     &model.Request{RequestID: id},
			Participant: p,
		})
		if err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM participants`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Errorf("participants: want 1, got %d", n)
	}
}

func TestGetByRequestIDNotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.GetByRequestID(context.Background(), "nope")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("want sql.ErrNoRows, got %v", err)
	}
}

func TestSaveNoSubEntities(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	in := &model.All{
		Request:     &model.Request{RequestID: "req-solo"},
		Participant: &model.Participant{INN: strp("1111111111")},
	}
	if _, err := s.Save(ctx, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	for _, tbl := range []string{"bank_accounts", "msp_info", "beneficiaries"} {
		var n int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM " + tbl).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", tbl, err)
		}
		if n != 0 {
			t.Errorf("%s: want 0, got %d", tbl, n)
		}
	}
}
