package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/middleware"
	"github.com/ramon/goals-tasks-api/internal/modules/goal/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/goal/usecase"
	"github.com/ramon/goals-tasks-api/internal/shared"
)

type GoalHandler struct {
	goalUseCase *usecase.GoalUseCase
}

func NewGoalHandler(goalUseCase *usecase.GoalUseCase) *GoalHandler {
	return &GoalHandler{
		goalUseCase: goalUseCase,
	}
}

func (h *GoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	var req dto.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	goal, err := h.goalUseCase.Create(r.Context(), userID, req)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusCreated, goal, "Meta criada com sucesso", traceID)
}

func (h *GoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	goalID := chi.URLParam(r, "id")
	if goalID == "" {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	var req dto.UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	goal, err := h.goalUseCase.Update(r.Context(), userID, goalID, req)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, goal, "Meta atualizada com sucesso", traceID)
}

func (h *GoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	goalID := chi.URLParam(r, "id")
	if goalID == "" {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	err := h.goalUseCase.Delete(r.Context(), userID, goalID)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, nil, "Meta excluída com sucesso", traceID)
}

func (h *GoalHandler) Get(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	goalID := chi.URLParam(r, "id")
	if goalID == "" {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	goal, err := h.goalUseCase.Get(r.Context(), userID, goalID)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, goal, "Meta recuperada com sucesso", traceID)
}

func (h *GoalHandler) List(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	activeOnly := r.URL.Query().Get("active")

	var result interface{}
	var err error

	if activeOnly == "true" {
		result, err = h.goalUseCase.ListActive(r.Context(), userID)
	} else {
		result, err = h.goalUseCase.ListAll(r.Context(), userID)
	}

	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, result, "Metas listadas com sucesso", traceID)
}
