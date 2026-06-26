package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/shared"
)

type AuthKey string

const UserIDContextKey AuthKey = "user_id"

func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(UserIDContextKey).(string); ok {
		return val
	}
	return ""
}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := shared.GetCorrelationID(r.Context())
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				shared.SendError(w, appErrors.ErrUnauthorized, traceID)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				shared.SendError(w, appErrors.ErrUnauthorized, traceID)
				return
			}

			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, appErrors.ErrUnauthorized
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				shared.SendError(w, appErrors.ErrUnauthorized, traceID)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				shared.SendError(w, appErrors.ErrUnauthorized, traceID)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok || userID == "" {
				shared.SendError(w, appErrors.ErrUnauthorized, traceID)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
