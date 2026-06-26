package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/middleware"
	"github.com/ramon/goals-tasks-api/internal/modules/task/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/task/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/task/usecase"
	"github.com/ramon/goals-tasks-api/internal/shared"
)

type TaskHandler struct {
	taskUseCase *usecase.TaskUseCase
}

func NewTaskHandler(taskUseCase *usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{
		taskUseCase: taskUseCase,
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	tasks, err := h.taskUseCase.Create(r.Context(), userID, req)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusCreated, tasks, "Tarefa(s) criada(s) com sucesso", traceID)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	task, err := h.taskUseCase.Update(r.Context(), userID, taskID, req)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, task, "Tarefa atualizada com sucesso", traceID)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	err := h.taskUseCase.Delete(r.Context(), userID, taskID)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, nil, "Tarefa excluída com sucesso", traceID)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		shared.SendError(w, appErrors.ErrBadRequest, traceID)
		return
	}

	task, err := h.taskUseCase.Get(r.Context(), userID, taskID)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, task, "Tarefa recuperada com sucesso", traceID)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	// 1. Completed history filter
	if r.URL.Query().Get("completed") == "true" {
		tasks, err := h.taskUseCase.ListCompletedHistory(r.Context(), userID)
		if err != nil {
			shared.SendError(w, err, traceID)
			return
		}
		shared.SendSuccess(w, http.StatusOK, tasks, "Histórico de tarefas concluídas recuperado", traceID)
		return
	}

	// 2. Period filter
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr != "" && endStr != "" {
		start, err1 := time.Parse(time.RFC3339, startStr)
		end, err2 := time.Parse(time.RFC3339, endStr)
		if err1 != nil || err2 != nil {
			shared.SendError(w, appErrors.NewAppError("INVALID_DATE_FORMAT", "as datas devem estar no formato RFC3339", 400), traceID)
			return
		}
		tasks, err := h.taskUseCase.ListByPeriod(r.Context(), userID, start, end)
		if err != nil {
			shared.SendError(w, err, traceID)
			return
		}
		shared.SendSuccess(w, http.StatusOK, tasks, "Tarefas do período recuperadas", traceID)
		return
	}

	// 3. Status filter
	statusStr := r.URL.Query().Get("status")
	if statusStr != "" {
		status := entity.TaskStatus(statusStr)
		tasks, err := h.taskUseCase.ListByStatus(r.Context(), userID, status)
		if err != nil {
			shared.SendError(w, err, traceID)
			return
		}
		shared.SendSuccess(w, http.StatusOK, tasks, "Tarefas filtradas por status", traceID)
		return
	}

	// Default: List tasks of current week
	tasks, err := h.taskUseCase.ListWeekly(r.Context(), userID)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}
	shared.SendSuccess(w, http.StatusOK, tasks, "Tarefas semanais recuperadas", traceID)
}

func (h *TaskHandler) ListWeekly(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	tasks, err := h.taskUseCase.ListWeekly(r.Context(), userID)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}
	shared.SendSuccess(w, http.StatusOK, tasks, "Tarefas semanais recuperadas", traceID)
}
