package shared

import (
	"context"
	"encoding/json"
	"net/http"
	
	"github.com/ramon/goals-tasks-api/internal/errors"
)

type CorrelationKey string

const CorrelationIDContextKey CorrelationKey = "correlation_id"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
	TraceID string      `json:"trace_id"`
}

type ErrorResponse struct {
	Success bool             `json:"success"`
	Error   *errors.AppError `json:"error"`
	TraceID string           `json:"trace_id"`
}

func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(CorrelationIDContextKey).(string); ok {
		return val
	}
	return ""
}

func SendSuccess(w http.ResponseWriter, status int, data interface{}, message string, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	resp := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
		TraceID: traceID,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func SendError(w http.ResponseWriter, err error, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	
	appErr := errors.MapError(err)
	w.WriteHeader(appErr.Status)
	
	resp := ErrorResponse{
		Success: false,
		Error:   appErr,
		TraceID: traceID,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
