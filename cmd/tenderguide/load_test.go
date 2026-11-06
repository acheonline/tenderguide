package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLoad(t *testing.T) {
	testdata, err := filepath.Abs("../../testdata")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	tests := []struct {
		name       string
		dir        string
		preload    bool
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "happy path",
			dir:        testdata,
			wantCode:   0,
			wantStdout: "OK: request testdata saved",
		},
		{
			name:       "duplicate",
			dir:        testdata,
			preload:    true,
			wantCode:   1,
			wantStderr: "duplicate request_id",
		},
		{
			name:       "nonexistent dir",
			dir:        "/nonexistent/path/xyz",
			wantCode:   1,
			wantStderr: "load:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "test.db")
			ctx := context.Background()

			if tt.preload {
				var out, errb bytes.Buffer
				if code := load(ctx, tt.dir, dbPath, &out, &errb); code != 0 {
					t.Fatalf("preload: code=%d stderr=%s", code, errb.String())
				}
			}

			var stdout, stderr bytes.Buffer
			code := load(ctx, tt.dir, dbPath, &stdout, &stderr)

			if code != tt.wantCode {
				t.Fatalf("code = %d, want %d (stderr=%s)", code, tt.wantCode, stderr.String())
			}
			if tt.wantStdout != "" && !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want contains %q", stdout.String(), tt.wantStdout)
			}
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want contains %q", stderr.String(), tt.wantStderr)
			}
			if tt.wantCode == 0 {
				if _, err := os.Stat(dbPath); err != nil {
					t.Errorf("db file not created: %v", err)
				}
			}
		})
	}
}

func TestRunLoadMissingDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runLoad(context.Background(), nil)
	_ = stdout
	_ = stderr
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
}

func TestRunLoadEmptyDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	var stdout, stderr bytes.Buffer
	code := load(context.Background(), dir, dbPath, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "no *.docx") {
		t.Errorf("stderr = %q, want contains 'no *.docx'", stderr.String())
	}
}
