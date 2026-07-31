package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	payload := taskResponse{
		ID:     1,
		Name:   "Test task",
		Status: task.StatusIncomplete,
	}

	err := writeJSON(recorder, http.StatusCreated, payload)
	if err != nil {
		t.Fatalf("writeJSON() error = %v", err)
	}

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}

	if got := response.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var got taskResponse

	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got != payload {
		t.Errorf("response = %#v, want %#v", got, payload)
	}
}

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := writeError(recorder, http.StatusBadRequest, "invalid request body")
	if err != nil {
		t.Fatalf("writeError() error = %v", err)
	}

	var got errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := errorResponse{
		Error: "invalid request body",
	}

	if got != want {
		t.Errorf("response = %#v, want %#v", got, want)
	}
}
