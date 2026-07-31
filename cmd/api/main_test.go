package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListenAddress(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		wantAddr string
	}{
		{
			name:     "default port",
			port:     "",
			wantAddr: ":8080",
		},
		{
			name:     "configured port",
			port:     "9000",
			wantAddr: ":9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)

			if got := listenAddress(); got != tt.wantAddr {
				t.Errorf("listenAddress() = %q, want %q", got, tt.wantAddr)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	server := newServer(":9000", handler)

	if server.Addr != ":9000" {
		t.Errorf("Addr = %q, want %q", server.Addr, ":9000")
	}

	if server.Handler == nil {
		t.Error("Handler = nil, want configured handler")
	}

	if server.ReadHeaderTimeout != readHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", server.ReadHeaderTimeout, readHeaderTimeout)
	}

	if server.ReadTimeout != readTimeout {
		t.Errorf("ReadTimeout = %v, want %v", server.ReadTimeout, readTimeout)
	}

	if server.WriteTimeout != writeTimeout {
		t.Errorf("WriteTimeout = %v, want %v", server.WriteTimeout, writeTimeout)
	}

	if server.IdleTimeout != idleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", server.IdleTimeout, idleTimeout)
	}
}

func TestNewApplication(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application := newApplication(logger)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "list tasks",
			method:     http.MethodGet,
			path:       "/tasks",
			wantStatus: http.StatusOK,
			wantBody:   "[]",
		},
		{
			name:       "unsupported method",
			method:     http.MethodPatch,
			path:       "/tasks",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			path:       "/missing",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()

			application.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}

			if tt.wantBody != "" {
				if got := strings.TrimSpace(recorder.Body.String()); got != tt.wantBody {
					t.Errorf("body = %q, want %q", got, tt.wantBody)
				}
			}
		})
	}
}
