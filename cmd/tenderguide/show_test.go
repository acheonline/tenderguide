package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunShow(t *testing.T) {
	testdata, err := filepath.Abs("../../testdata")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	tests := []struct {
		name       string
		request    string
		format     string
		preload    bool
		wantCode   int
		wantStdout string
		wantStderr string
		wantJSON   bool
	}{
		{
			name:       "happy path text",
			request:    "testdata",
			format:     "text",
			preload:    true,
			wantCode:   0,
			wantStdout: "=== Заявка ===",
		},
		{
			name:     "happy path json",
			request:  "testdata",
			format:   "json",
			preload:  true,
			wantCode: 0,
			wantJSON: true,
		},
		{
			name:       "not found",
			request:    "nope",
			format:     "text",
			preload:    true,
			wantCode:   1,
			wantStderr: "request not found",
		},
		{
			name:       "invalid format",
			request:    "testdata",
			format:     "xml",
			preload:    true,
			wantCode:   2,
			wantStderr: "invalid format",
		},
		{
			name:       "missing request",
			request:    "",
			format:     "text",
			wantCode:   2,
			wantStderr: "--request is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "test.db")
			ctx := context.Background()

			if tt.preload {
				var out, errb bytes.Buffer
				if code := load(ctx, testdata, dbPath, &out, &errb); code != 0 {
					t.Fatalf("preload: code=%d stderr=%s", code, errb.String())
				}
			}

			var stdout, stderr bytes.Buffer
			code := runShow(ctx, []string{"--request", tt.request, "--db", dbPath, "--format", tt.format})
			_ = stdout
			_ = stderr

			if code != tt.wantCode {
				t.Fatalf("code = %d, want %d", code, tt.wantCode)
			}
		})
	}
}

func TestShow(t *testing.T) {
	testdata, err := filepath.Abs("../../testdata")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	tests := []struct {
		name       string
		request    string
		format     string
		preload    bool
		wantCode   int
		wantStdout string
		wantStderr string
		wantJSON   bool
	}{
		{
			name:       "happy path text",
			request:    "testdata",
			format:     "text",
			preload:    true,
			wantCode:   0,
			wantStdout: "=== Заявка ===",
		},
		{
			name:     "happy path json",
			request:  "testdata",
			format:   "json",
			preload:  true,
			wantCode: 0,
			wantJSON: true,
		},
		{
			name:       "not found",
			request:    "nope",
			format:     "text",
			preload:    true,
			wantCode:   1,
			wantStderr: "request not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "test.db")
			ctx := context.Background()

			if tt.preload {
				var out, errb bytes.Buffer
				if code := load(ctx, testdata, dbPath, &out, &errb); code != 0 {
					t.Fatalf("preload: code=%d stderr=%s", code, errb.String())
				}
			}

			var stdout, stderr bytes.Buffer
			code := show(ctx, tt.request, dbPath, tt.format, &stdout, &stderr)

			if code != tt.wantCode {
				t.Fatalf("code = %d, want %d (stderr=%s)", code, tt.wantCode, stderr.String())
			}
			if tt.wantStdout != "" && !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want contains %q", stdout.String(), tt.wantStdout)
			}
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want contains %q", stderr.String(), tt.wantStderr)
			}
			if tt.wantJSON {
				var v any
				if err := json.Unmarshal(stdout.Bytes(), &v); err != nil {
					t.Errorf("stdout is not valid JSON: %v", err)
				}
			}
		})
	}
}
