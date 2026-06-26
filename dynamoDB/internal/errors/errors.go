package errors

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

var (
	ErrUserNotFound       = NewAppError("USER_NOT_FOUND", "Usuário não encontrado", http.StatusNotFound)
	ErrUserAlreadyExists  = NewAppError("USER_ALREADY_EXISTS", "Usuário já cadastrado", http.StatusConflict)
	ErrInvalidCredentials = NewAppError("INVALID_CREDENTIALS", "Credenciais inválidas", http.StatusUnauthorized)
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", "Não autorizado", http.StatusUnauthorized)
	ErrForbidden          = NewAppError("FORBIDDEN", "Acesso proibido", http.StatusForbidden)
	
	ErrGoalNotFound       = NewAppError("GOAL_NOT_FOUND", "Meta não encontrada", http.StatusNotFound)
	ErrGoalInvalidDates   = NewAppError("GOAL_INVALID_DATES", "A data de início deve ser anterior ou igual à data de fim", http.StatusBadRequest)
	
	ErrTaskNotFound       = NewAppError("TASK_NOT_FOUND", "Tarefa não encontrada", http.StatusNotFound)
	ErrInvalidTaskStatus  = NewAppError("INVALID_TASK_STATUS", "Status de tarefa inválido", http.StatusBadRequest)
	
	ErrInternalServer     = NewAppError("INTERNAL_SERVER_ERROR", "Erro interno no servidor", http.StatusInternalServerError)
	ErrBadRequest         = NewAppError("BAD_REQUEST", "Requisição inválida", http.StatusBadRequest)
	ErrRateLimitExceeded  = NewAppError("RATE_LIMIT_EXCEEDED", "Limite de requisições excedido", http.StatusTooManyRequests)
)

func MapError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternalServer
}
