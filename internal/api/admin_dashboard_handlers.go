package api

import (
	"log/slog"
	"net/http"

	"night-owls-go/internal/service"
)

// AdminDashboardHandler handles admin dashboard operations
type AdminDashboardHandler struct {
	dashboardService *service.AdminDashboardService
	logger           *slog.Logger
}

// NewAdminDashboardHandler creates a new AdminDashboardHandler
func NewAdminDashboardHandler(dashboardService *service.AdminDashboardService, logger *slog.Logger) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		dashboardService: dashboardService,
		logger:           logger.With("handler", "AdminDashboardHandler"),
	}
}

// GetDashboardHandler handles GET /api/admin/dashboard
// @Summary Get admin dashboard metrics
// @Description Returns comprehensive dashboard metrics including unfilled shifts, member contributions, and quality metrics
// @Tags admin-dashboard
// @Produce json
// @Success 200 {object} service.AdminDashboard "Dashboard metrics"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/dashboard [get]
func (h *AdminDashboardHandler) GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.dashboardService.GetDashboard(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate dashboard metrics", h.logger, "error", err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, dashboard, h.logger)
} 