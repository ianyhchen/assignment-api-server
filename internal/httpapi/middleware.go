package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type statusResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (writer *statusResponseWriter) WriteHeader(status int) {
	if writer.wroteHeader {
		return
	}

	writer.status = status
	writer.wroteHeader = true
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *statusResponseWriter) Write(body []byte) (int, error) {
	if !writer.wroteHeader {
		writer.WriteHeader(http.StatusOK)
	}

	return writer.ResponseWriter.Write(body)
}

func (writer *statusResponseWriter) Status() int {
	if writer.status == 0 {
		return http.StatusOK
	}

	return writer.status
}

func (writer *statusResponseWriter) Unwrap() http.ResponseWriter {
	return writer.ResponseWriter
}

// RequestLogger logs completed HTTP requests.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := &statusResponseWriter{ResponseWriter: w}
			start := time.Now()

			defer func() {
				logger.InfoContext(
					r.Context(),
					"request completed",
					"method",
					r.Method,
					"path",
					r.URL.Path,
					"status",
					writer.Status(),
					"duration",
					time.Since(start),
				)
			}()
			next.ServeHTTP(writer, r)
		})
	}
}

// Recoverer converts handler panics into internal server error responses.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				logger.ErrorContext(
					r.Context(),
					"panic recovered",
					"panic",
					recovered,
					"stack",
					string(debug.Stack()),
				)

				if err := writeError(w, http.StatusInternalServerError, "internal server error"); err != nil {
					logger.ErrorContext(
						r.Context(),
						"write panic response failed",
						"error",
						err,
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
