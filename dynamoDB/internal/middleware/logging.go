package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ramon/goals-tasks-api/internal/shared"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		traceID := shared.GetCorrelationID(r.Context())

		slog.InfoContext(r.Context(), "HTTP Request",
			"trace_id", traceID,
			"request_id", traceID,
			"method", r.Method,
			"route", r.URL.Path,
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"status_code", wrapper.statusCode,
		)
	})
}
