package loader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFullFixture(t *testing.T) {
	inputs, err := Load("../../testdata")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(inputs) != 9 {
		t.Fatalf("want 9 inputs, got %d", len(inputs))
	}

	for i := 1; i < len(inputs); i++ {
		if inputs[i-1].FileName > inputs[i].FileName {
			t.Fatalf("inputs not sorted: %q > %q", inputs[i-1].FileName, inputs[i].FileName)
		}
	}

	var form1Paras []string
	var form2Tables int
	for _, in := range inputs {
		base := filepath.Base(in.FileName)
		if strings.HasPrefix(base, "Форма 1 ") {
			form1Paras = in.Paragraphs
		}
		if strings.HasPrefix(base, "Форма 2 ") {
			form2Tables = len(in.Tables)
		}
	}

	if form1Paras == nil {
		t.Fatal("Форма 1 не найдена среди inputs")
	}
	foundNumber := false
	for _, p := range form1Paras {
		if strings.Contains(p, "32616122324") {
			foundNumber = true
			break
		}
	}
	if !foundNumber {
		t.Fatalf("в Форма 1 не найден номер процедуры; абзацы: %v", form1Paras)
	}

	if form2Tables == 0 {
		t.Fatal("в Форма 2 не найдено таблиц")
	}
}

func TestLoadFileNameRelative(t *testing.T) {
	inputs, err := Load("../../testdata")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, in := range inputs {
		if filepath.IsAbs(in.FileName) {
			t.Fatalf("FileName is absolute: %q", in.FileName)
		}
		if strings.Contains(in.FileName, "..") {
			t.Fatalf("FileName contains '..': %q", in.FileName)
		}
	}
}

func TestLoadNoDocx(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("want error for dir without *.docx")
	}
}

func TestLoadNotExist(t *testing.T) {
	if _, err := Load("/nonexistent/path/xyz"); err == nil {
		t.Fatal("want error for nonexistent dir")
	}
}
