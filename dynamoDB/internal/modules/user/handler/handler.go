package handler

import (
	"encoding/json"
	"net/http"

	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/modules/user/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/user/usecase"
	"github.com/ramon/goals-tasks-api/internal/shared"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	user, err := h.authUseCase.Register(r.Context(), req)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusCreated, user, "Usuário cadastrado com sucesso", traceID)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	resp, err := h.authUseCase.Login(r.Context(), req)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, resp, "Login realizado com sucesso", traceID)
}
