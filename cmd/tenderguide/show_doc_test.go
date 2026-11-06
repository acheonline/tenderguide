package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acheonline/tenderguide/internal/model"
	"github.com/acheonline/tenderguide/internal/storage"
)

func strp(s string) *string { return &s }
func boolp(b bool) *bool    { return &b }

func setupDocStore(t *testing.T) (dbPath string) {
	t.Helper()
	dbPath = filepath.Join(t.TempDir(), "test.db")
	ctx := context.Background()
	store, err := storage.Open(ctx, "file:"+dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer store.Close()

	doc := &model.DocAll{
		Info: &model.DocInfo{
			DocID:                        "testdata-doc",
			CustomerName:                 strp("АО «ТЭК СПб»"),
			CustomerINN:                  strp("7838129301"),
			OrganizerName:                strp("СПб ГКУ «Управление заказчика»"),
			ProcedureNumber:              strp("32616122324"),
			ProcedureName:                strp("Оказание услуг по обеспечению безопасности веб-приложений"),
			NMCK:                         strp("14414600.00"),
			VAT:                          strp("2013000.00"),
			SubmissionDeadline:           strp("2026-07-03T09:00:00Z"),
			ReviewDeadline:               strp("2026-07-16T20:59:00Z"),
			BidSecurityAmount:            strp("720730.00"),
			BidSecurityPercent:           strp("5"),
			BidSecurityType:              strp("cash_or_bank_guarantee"),
			ContractSecurityRequired:     boolp(false),
			PriceCriterionWeight:         strp("70"),
			QualificationCriterionWeight: strp("30"),
			LawBasis:                     strp("223-ФЗ"),
			CreatedAt:                    "2026-11-06T15:00:00Z",
		},
		Requirements: []*model.DocRequirement{
			{Kind: "requirement", Code: strp("6.1.2"), Name: strp("непроведение ликвидации участника закупки"), Required: boolp(true), SourceSection: strp("6.1")},
			{Kind: "requirement", Code: strp("6.1.10"), Name: strp("конфликта интересов"), Required: boolp(true), SourceSection: strp("6.1")},
			{Kind: "document", Code: strp("7.1.2"), Name: strp("Копию выписки из Единого государственного реестра юридических лиц"), Required: boolp(true), SourceSection: strp("7.1")},
			{Kind: "document", Code: strp("форма 2"), Name: strp("Анкета Участника"), Required: boolp(true), SourceSection: strp("8.1.1")},
			{Kind: "document", Code: strp("форма 11"), Name: strp("Опыт участника конкурса"), Required: boolp(false), SourceSection: strp("8.1.1")},
		},
	}
	if _, err := store.SaveDoc(ctx, doc); err != nil {
		t.Fatalf("SaveDoc: %v", err)
	}
	return dbPath
}

func TestShowDocText(t *testing.T) {
	dbPath := setupDocStore(t)
	code := runShowDoc(context.Background(), []string{"--id", "testdata-doc", "--db", dbPath})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestShowDocJSON(t *testing.T) {
	dbPath := setupDocStore(t)
	code := runShowDoc(context.Background(), []string{"--id", "testdata-doc", "--db", dbPath, "--format", "json"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestShowDocNotFound(t *testing.T) {
	dbPath := setupDocStore(t)
	code := runShowDoc(context.Background(), []string{"--id", "nope", "--db", dbPath})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestShowDocNoID(t *testing.T) {
	code := runShowDoc(context.Background(), []string{"--db", "/tmp/x.db"})
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestShowDocInvalidFormat(t *testing.T) {
	code := runShowDoc(context.Background(), []string{"--id", "x", "--format", "xml"})
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestShowDocTextContainsSections(t *testing.T) {
	dbPath := setupDocStore(t)
	// Перехватываем stdout через перенаправление os.Stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runShowDoc(context.Background(), []string{"--id", "testdata-doc", "--db", dbPath})

	_ = w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	for _, want := range []string{
		"=== Документация о закупке ===",
		"=== Требования к участнику",
		"=== Документы в составе заявки",
		"14414600.00",
		"720730.00",
		"223-ФЗ",
		"6.1.2",
		"выписки из Единого государственного реестра",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not contain %q", want)
		}
	}
}

func TestShowDocJSONValid(t *testing.T) {
	dbPath := setupDocStore(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runShowDoc(context.Background(), []string{"--id", "testdata-doc", "--db", dbPath, "--format", "json"})

	_ = w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	buf := make([]byte, 128*1024)
	n, _ := r.Read(buf)
	var v map[string]any
	if err := json.Unmarshal(buf[:n], &v); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := v["Info"]; !ok {
		t.Error("JSON missing Info")
	}
	if _, ok := v["Requirements"]; !ok {
		t.Error("JSON missing Requirements")
	}
}
