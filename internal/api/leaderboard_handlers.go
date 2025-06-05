package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"night-owls-go/internal/service"
)

// LeaderboardHandler handles leaderboard and points-related operations
type LeaderboardHandler struct {
	pointsService *service.PointsService
	logger        *slog.Logger
}

// NewLeaderboardHandler creates a new LeaderboardHandler
func NewLeaderboardHandler(pointsService *service.PointsService, logger *slog.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{
		pointsService: pointsService,
		logger:        logger.With("handler", "LeaderboardHandler"),
	}
}

// GetLeaderboardHandler handles GET /api/leaderboard
// @Summary Get points leaderboard
// @Description Returns the top users ranked by total points
// @Tags leaderboard
// @Produce json
// @Param limit query int false "Number of users to return" default(10)
// @Success 200 {array} LeaderboardEntry "Leaderboard entries"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/leaderboard [get]
func (h *LeaderboardHandler) GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := int32(10) // default
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = int32(parsedLimit)
		}
	}

	entries, err := h.pointsService.GetLeaderboard(r.Context(), limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get leaderboard", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get leaderboard", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	response := make([]LeaderboardEntry, len(entries))
	for i, entry := range entries {
		name := ""
		if entry.Name.Valid {
			name = entry.Name.String
		}

		totalPoints := int64(0)
		if entry.TotalPoints.Valid {
			totalPoints = entry.TotalPoints.Int64
		}

		shiftCount := int64(0)
		if entry.ShiftCount.Valid {
			shiftCount = entry.ShiftCount.Int64
		}

		response[i] = LeaderboardEntry{
			UserID:           entry.UserID,
			Name:             name,
			TotalPoints:      totalPoints,
			ShiftCount:       shiftCount,
			AchievementCount: entry.AchievementCount,
			ActivityStatus:   entry.ActivityStatus,
		}
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetStreakLeaderboardHandler handles GET /api/leaderboard/shifts
// @Summary Get shift leaderboard
// @Description Returns the top users ranked by shift count
// @Tags leaderboard
// @Produce json
// @Param limit query int false "Number of users to return" default(10)
// @Success 200 {array} LeaderboardEntry "Shift leaderboard entries"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/leaderboard/shifts [get]
func (h *LeaderboardHandler) GetStreakLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := int32(10) // default
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = int32(parsedLimit)
		}
	}

	entries, err := h.pointsService.GetShiftLeaderboard(r.Context(), limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get shift leaderboard", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get shift leaderboard", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	response := make([]LeaderboardEntry, len(entries))
	for i, entry := range entries {
		name := ""
		if entry.Name.Valid {
			name = entry.Name.String
		}

		totalPoints := int64(0)
		if entry.TotalPoints.Valid {
			totalPoints = entry.TotalPoints.Int64
		}

		shiftCount := int64(0)
		if entry.ShiftCount.Valid {
			shiftCount = entry.ShiftCount.Int64
		}

		response[i] = LeaderboardEntry{
			UserID:           entry.UserID,
			Name:             name,
			TotalPoints:      totalPoints,
			ShiftCount:       shiftCount,
			AchievementCount: entry.AchievementCount,
			ActivityStatus:   entry.ActivityStatus,
		}
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetUserStatsHandler handles GET /api/user/stats
// @Summary Get current user's points and achievements
// @Description Returns the authenticated user's complete statistics
// @Tags user-stats
// @Produce json
// @Success 200 {object} UserStatsResponse "User statistics"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/user/stats [get]
func (h *LeaderboardHandler) GetUserStatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok || userID == 0 {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	// Get user stats
	stats, err := h.pointsService.GetUserStats(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get user stats", "user_id", userID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user stats", h.logger, "error", err.Error())
		return
	}

	// Get user rank
	rank, err := h.pointsService.GetUserRank(r.Context(), userID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "Failed to get user rank", "user_id", userID, "error", err)
		rank = 0 // Default if rank calculation fails
	}

	// Convert to API response format
	var lastActivityDate *string
	if stats.LastActivityDate.Valid {
		dateStr := stats.LastActivityDate.Time.Format("2006-01-02")
		lastActivityDate = &dateStr
	}

	name := ""
	if stats.Name.Valid {
		name = stats.Name.String
	}

	response := UserStatsResponse{
		UserID:           stats.UserID,
		Name:             name,
		TotalPoints:      stats.TotalPoints.Int64,
		ShiftCount:       stats.ShiftCount.Int64,
		LastActivityDate: lastActivityDate,
		Rank:             rank,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetUserPointsHistoryHandler handles GET /api/user/points/history
// @Summary Get current user's points history
// @Description Returns the authenticated user's recent points transactions
// @Tags user-stats
// @Produce json
// @Param limit query int false "Number of entries to return" default(20)
// @Success 200 {array} PointsHistoryEntry "Points history entries"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/user/points/history [get]
func (h *LeaderboardHandler) GetUserPointsHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok || userID == 0 {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := int64(20) // default
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	history, err := h.pointsService.GetUserPointsHistory(r.Context(), userID, limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get user points history", "user_id", userID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get points history", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	response := make([]PointsHistoryEntry, len(history))
	for i, entry := range history {
		var shiftStart *time.Time
		if entry.ShiftStart.Valid {
			shiftStart = &entry.ShiftStart.Time
		}

		createdAt := time.Time{}
		if entry.CreatedAt.Valid {
			createdAt = entry.CreatedAt.Time
		}

		response[i] = PointsHistoryEntry{
			PointsAwarded: entry.PointsAwarded,
			Reason:        entry.Reason,
			Multiplier:    entry.Multiplier.Float64,
			CreatedAt:     createdAt,
			ShiftStart:    shiftStart,
		}
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetUserAchievementsHandler handles GET /api/user/achievements
// @Summary Get current user's achievements
// @Description Returns the authenticated user's earned achievements
// @Tags user-stats
// @Produce json
// @Success 200 {array} AchievementResponse "User achievements"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/user/achievements [get]
func (h *LeaderboardHandler) GetUserAchievementsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok || userID == 0 {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	achievements, err := h.pointsService.GetUserAchievements(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get user achievements", "user_id", userID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get achievements", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	response := make([]AchievementResponse, len(achievements))
	for i, achievement := range achievements {
		var earnedAt *time.Time
		if achievement.EarnedAt.Valid {
			earnedAt = &achievement.EarnedAt.Time
		}

		response[i] = AchievementResponse{
			AchievementID: achievement.AchievementID,
			Name:          achievement.Name,
			Description:   achievement.Description,
			Icon:          achievement.Icon.String,
			EarnedAt:      earnedAt,
		}
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetAvailableAchievementsHandler handles GET /api/user/achievements/available
// @Summary Get available achievements for current user
// @Description Returns achievements the authenticated user hasn't earned yet
// @Tags user-stats
// @Produce json
// @Success 200 {array} AchievementResponse "Available achievements"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/user/achievements/available [get]
func (h *LeaderboardHandler) GetAvailableAchievementsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok || userID == 0 {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	achievements, err := h.pointsService.GetAvailableAchievements(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get available achievements", "user_id", userID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get available achievements", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	response := make([]AchievementResponse, len(achievements))
	for i, achievement := range achievements {
		var shiftsThreshold *int64
		if achievement.ShiftsThreshold.Valid {
			shiftsThreshold = &achievement.ShiftsThreshold.Int64
		}

		response[i] = AchievementResponse{
			AchievementID:   achievement.AchievementID,
			Name:            achievement.Name,
			Description:     achievement.Description,
			Icon:            achievement.Icon.String,
			ShiftsThreshold: shiftsThreshold,
		}
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetActivityFeedHandler handles GET /api/leaderboard/activity
// @Summary Get recent community activity
// @Description Returns recent point-earning activities across all users
// @Tags leaderboard
// @Produce json
// @Param limit query int false "Number of activities to return" default(20)
// @Success 200 {array} ActivityFeedEntry "Activity feed entries"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/leaderboard/activity [get]
func (h *LeaderboardHandler) GetActivityFeedHandler(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := int32(20) // default
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = int32(parsedLimit)
		}
	}

	activities, err := h.pointsService.GetRecentActivity(r.Context(), limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get activity feed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get activity feed", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	response := make([]ActivityFeedEntry, len(activities))
	for i, activity := range activities {
		name := ""
		if activity.Name.Valid {
			name = activity.Name.String
		}

		createdAt := time.Time{}
		if activity.CreatedAt.Valid {
			createdAt = activity.CreatedAt.Time
		}

		response[i] = ActivityFeedEntry{
			UserName:      name,
			PointsAwarded: activity.PointsAwarded,
			Reason:        activity.Reason,
			ActivityType:  activity.ActivityType,
			CreatedAt:     createdAt,
		}
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}
