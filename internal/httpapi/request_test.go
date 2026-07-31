package httpapi

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantError bool
	}{
		{
			name: "valid request",
			body: `{"name":"Task","status":0}`,
		},
		{
			name:      "empty body",
			body:      "",
			wantError: true,
		},
		{
			name:      "malformed JSON",
			body:      `{"name":`,
			wantError: true,
		},
		{
			name:      "unknown field",
			body:      `{"name":"Task","status":0,"other":true}`,
			wantError: true,
		},
		{
			name: "trailing whitespace",
			body: `{"name":"Task","status":0}` + "\n\t ",
		},
		{
			name: "explicit status zero",
			body: `{"name":"Task","status":0}`,
		},
		{
			name:      "multiple JSON values",
			body:      `{"name":"Task","status":0} {"name":"Other","status":1}`,
			wantError: true,
		},
		{
			name:      "wrong field type",
			body:      `{"name":"Task","status":"incomplete"}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/tasks", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			var got createTaskRequest

			err := decodeJSON(recorder, request, &got)

			if tt.wantError && err == nil {
				t.Fatal("decodeJSON() error = nil, want error")
			}

			if !tt.wantError && err != nil {
				t.Fatalf("decodeJSON() error = %v, want nil", err)
			}
		})
	}
}

func TestDecodeJSONPreservesExplicitZeroStatus(t *testing.T) {
	request := httptest.NewRequest("POST", "/tasks", strings.NewReader(`{"name":"Task","status":0}`))
	recorder := httptest.NewRecorder()

	var got createTaskRequest

	if err := decodeJSON(recorder, request, &got); err != nil {
		t.Fatalf("decodeJSON() error = %v", err)
	}

	if got.Name == nil {
		t.Fatal("decoded name = nil, want provided name")
	}

	if got.Status == nil {
		t.Fatal("decoded status = nil, want provided status")
	}

	if *got.Name != "Task" {
		t.Errorf("decoded name = %q, want %q", *got.Name, "Task")
	}

	if *got.Status != 0 {
		t.Errorf("decoded status = %d, want 0", *got.Status)
	}
}

func TestDecodeJSONRejectsLargeBody(t *testing.T) {
	body := `{"name":"` +
		strings.Repeat("a", int(maxRequestBodyBytes)) +
		`","status":0}`

	request := httptest.NewRequest("POST", "/tasks", strings.NewReader(body))
	recorder := httptest.NewRecorder()

	var got createTaskRequest

	if err := decodeJSON(recorder, request, &got); err == nil {
		t.Fatal("decodeJSON() error = nil, want body-size error")
	}
}

func TestParseTaskID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantID    uint64
		wantError bool
	}{
		{
			name:   "valid ID",
			value:  "42",
			wantID: 42,
		},
		{
			name:      "empty ID",
			value:     "",
			wantError: true,
		},
		{
			name:      "zero ID",
			value:     "0",
			wantError: true,
		},
		{
			name:      "negative ID",
			value:     "-1",
			wantError: true,
		},
		{
			name:      "non-numeric ID",
			value:     "abc",
			wantError: true,
		},
		{
			name:      "ID overflows uint64",
			value:     "18446744073709551616",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parseTaskID(tt.value)

			if tt.wantError {
				if !errors.Is(err, task.ErrInvalidTaskID) {
					t.Errorf("parseTaskID(%q) error = %v, want ErrInvalidTaskID", tt.value, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseTaskID(%q) error = %v", tt.value, err)
			}

			if id != tt.wantID {
				t.Errorf("parseTaskID(%q) = %d, want %d", tt.value, id, tt.wantID)
			}
		})
	}
}
