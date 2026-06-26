package middleware

import (
	"context"
	"net/http"

	"github.com/ramon/goals-tasks-api/internal/shared"
	"github.com/ramon/goals-tasks-api/internal/utils"
)

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corID := r.Header.Get("X-Correlation-ID")
		if corID == "" {
			corID = utils.GenerateUUID()
		}

		ctx := context.WithValue(r.Context(), shared.CorrelationIDContextKey, corID)
		w.Header().Set("X-Correlation-ID", corID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
