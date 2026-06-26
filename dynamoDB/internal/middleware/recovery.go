package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/shared"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				traceID := shared.GetCorrelationID(r.Context())
				
				slog.ErrorContext(r.Context(), "Servidor recuperado de pânico",
					"trace_id", traceID,
					"error", fmt.Sprintf("%v", err),
					"stack", string(debug.Stack()),
				)

				shared.SendError(w, appErrors.ErrInternalServer, traceID)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
