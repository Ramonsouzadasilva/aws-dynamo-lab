package handler

import (
	"net/http"

	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/middleware"
	"github.com/ramon/goals-tasks-api/internal/modules/dashboard/usecase"
	"github.com/ramon/goals-tasks-api/internal/shared"
)

type DashboardHandler struct {
	dashboardUseCase *usecase.DashboardUseCase
}

func NewDashboardHandler(dashboardUseCase *usecase.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{
		dashboardUseCase: dashboardUseCase,
	}
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	traceID := shared.GetCorrelationID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		shared.SendError(w, appErrors.ErrUnauthorized, traceID)
		return
	}

	dashboard, err := h.dashboardUseCase.GetDashboard(r.Context(), userID)
	if err != nil {
		shared.SendError(w, err, traceID)
		return
	}

	shared.SendSuccess(w, http.StatusOK, dashboard, "Dashboard recuperado com sucesso", traceID)
}
