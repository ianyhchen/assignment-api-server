package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusResponseWriterRecordsStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &statusResponseWriter{
		ResponseWriter: recorder,
	}

	writer.WriteHeader(http.StatusCreated)
	writer.WriteHeader(http.StatusBadRequest)

	if writer.Status() != http.StatusCreated {
		t.Errorf("status = %d, want %d", writer.Status(), http.StatusCreated)
	}

	if recorder.Code != http.StatusCreated {
		t.Errorf("response status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestStatusResponseWriterDefaultsToOK(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &statusResponseWriter{
		ResponseWriter: recorder,
	}

	if _, err := writer.Write([]byte("response")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if writer.Status() != http.StatusOK {
		t.Errorf("status = %d, want %d", writer.Status(), http.StatusOK)
	}
}

func TestRequestLogger(t *testing.T) {
	var output bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&output, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := RequestLogger(logger)(next)
	request := httptest.NewRequest(http.MethodPost, "/tasks", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	logged := output.String()

	for _, want := range []string{
		"request completed",
		"method=POST",
		"path=/tasks",
		"status=201",
		"duration=",
	} {
		if !strings.Contains(logged, want) {
			t.Errorf("log = %q, want to contain %q", logged, want)
		}
	}
}

func TestRecoverer(t *testing.T) {
	var output bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&output, nil))

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("database password was rejected")
	})

	handler := Recoverer(logger)(next)
	request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	body := recorder.Body.Bytes()

	var got errorResponse

	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Error != "internal server error" {
		t.Errorf("error = %q, want %q", got.Error, "internal server error")
	}

	if strings.Contains(string(body), "database password") {
		t.Error("response exposed the recovered panic")
	}

	if !strings.Contains(output.String(), "database password was rejected") {
		t.Error("log does not contain the recovered panic")
	}
}

func TestRequestLoggerRecordsRecoveredPanicStatus(t *testing.T) {
	var output bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&output, nil))

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("unexpected failure")
	})

	handler := Recoverer(logger)(next)
	handler = RequestLogger(logger)(handler)

	request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if !strings.Contains(output.String(), "status=500") {
		t.Errorf("log = %q, want status=500", output.String())
	}
}
