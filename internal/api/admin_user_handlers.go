package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/go-chi/chi/v5"
)

// Request/response types
type createUserRequest struct {
	Phone string  `json:"phone"`
	Name  string  `json:"name"`
	Role  *string `json:"role,omitempty"` // Optional role
}

type updateUserRequest struct {
	Phone string  `json:"phone"`
	Name  string  `json:"name"`
	Role  *string `json:"role,omitempty"` // Optional role, assuming we'll add update role logic later
}

// UserAPIResponse defines the structure for user data sent to the frontend.
type UserAPIResponse struct {
	ID        int64   `json:"id"`
	Phone     string  `json:"phone"`
	Name      *string `json:"name"`
	CreatedAt string  `json:"created_at"`
	Role      string  `json:"role"` // Added role
}

// AdminUserHandler handles admin-specific user API requests.
type AdminUserHandler struct {
	db      db.Querier
	logger *slog.Logger
}

// NewAdminUserHandler creates a new AdminUserHandler.
func NewAdminUserHandler(db db.Querier, logger *slog.Logger) *AdminUserHandler {
	return &AdminUserHandler{
		db:      db,
			logger: logger.With("handler", "AdminUserHandler"),
	}
}

// AdminListUsers handles GET /api/admin/users
// @Summary List all users (Admin)
// @Description Get a list of all users in the system. Requires admin authentication.
// @Tags admin/users
// @Produce json
// @Success 200 {array} UserAPIResponse "List of users"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/users [get]
func (h *AdminUserHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	dbUsers, err := h.db.ListUsers(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to list users", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users", h.logger)
		return
	}

	apiUsers := make([]UserAPIResponse, 0, len(dbUsers))
	for _, u := range dbUsers {
		var namePtr *string
		if u.Name.Valid {
			namePtr = &u.Name.String
		}
		var createdAtStr string
		if u.CreatedAt.Valid {
			createdAtStr = u.CreatedAt.Time.Format(time.RFC3339)
		}
		apiUsers = append(apiUsers, UserAPIResponse{
			ID:        u.UserID,
			Phone:     u.Phone,
			Name:      namePtr,
			CreatedAt: createdAtStr,
			Role:      u.Role, // Added role
		})
	}

	RespondWithJSON(w, http.StatusOK, apiUsers, h.logger)
}

// AdminGetUser handles GET /api/admin/users/{id}
// @Summary Get a user by ID (Admin)
// @Description Get a specific user's details by their ID. Requires admin authentication.
// @Tags admin/users
// @Produce json
// @Param id path int64 true "User ID"
// @Success 200 {object} UserAPIResponse "User details"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/users/{id} [get]
func (h *AdminUserHandler) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger, "error", err)
		return
	}

	dbUser, err := h.db.GetUserByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
		} else {
			h.logger.ErrorContext(r.Context(), "Failed to get user by ID", "user_id", id, "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user", h.logger)
		}
		return
	}

	var namePtr *string
	if dbUser.Name.Valid {
		namePtr = &dbUser.Name.String
	}
	var createdAtStr string
	if dbUser.CreatedAt.Valid {
		createdAtStr = dbUser.CreatedAt.Time.Format(time.RFC3339)
	}

	apiUser := UserAPIResponse{
		ID:        dbUser.UserID,
		Phone:     dbUser.Phone,
		Name:      namePtr,
		CreatedAt: createdAtStr,
		Role:      dbUser.Role, // Added role
	}

	RespondWithJSON(w, http.StatusOK, apiUser, h.logger)
}

// AdminCreateUser handles POST /api/admin/users
// @Summary Create a new user (Admin)
// @Description Create a new user in the system. Requires admin authentication.
// @Tags admin/users
// @Accept json
// @Produce json
// @Param user body createUserRequest true "User information"
// @Success 201 {object} UserAPIResponse "Created user"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/users [post]
func (h *AdminUserHandler) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger, "error", err)
		return
	}

	// Validate phone number
	if req.Phone == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number is required", h.logger)
		return
	}

	// Check if user with this phone already exists
	_, err := h.db.GetUserByPhone(r.Context(), req.Phone)
	if err == nil {
		// User exists
		RespondWithError(w, http.StatusBadRequest, "User with this phone number already exists", h.logger)
		return
	} else if err != sql.ErrNoRows {
		// Database error
		h.logger.ErrorContext(r.Context(), "Failed to check for existing user", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user", h.logger)
		return
	}

	// Validate role if provided
	if req.Role != nil && !isValidRole(*req.Role) {
		RespondWithError(w, http.StatusBadRequest, "Invalid role specified. Must be one of: admin, owl, guest", h.logger)
		return
	}

	// Create user
	params := db.CreateUserParams{
		Phone: req.Phone,
		Name:  sql.NullString{String: req.Name, Valid: req.Name != ""},
		Role:  sql.NullString{String: derefString(req.Role), Valid: req.Role != nil},
	}

	dbUser, err := h.db.CreateUser(r.Context(), params)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to create user", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user", h.logger)
		return
	}

	var namePtr *string
	if dbUser.Name.Valid {
		namePtr = &dbUser.Name.String
	}
	var createdAtStr string
	if dbUser.CreatedAt.Valid {
		createdAtStr = dbUser.CreatedAt.Time.Format(time.RFC3339)
	}

	apiUser := UserAPIResponse{
		ID:        dbUser.UserID,
		Phone:     dbUser.Phone,
		Name:      namePtr,
		CreatedAt: createdAtStr,
		Role:      dbUser.Role, // Added role
	}

	RespondWithJSON(w, http.StatusCreated, apiUser, h.logger)
}

// AdminUpdateUser handles PUT /api/admin/users/{id}
// @Summary Update a user (Admin)
// @Description Update a user's details by their ID. Requires admin authentication.
// @Tags admin/users
// @Accept json
// @Produce json
// @Param id path int64 true "User ID"
// @Param user body updateUserRequest true "User information"
// @Success 200 {object} UserAPIResponse "Updated user"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/users/{id} [put]
func (h *AdminUserHandler) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger, "error", err)
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger, "error", err)
		return
	}

	// Check if user exists
	_, err = h.db.GetUserByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
		} else {
			h.logger.ErrorContext(r.Context(), "Failed to get user by ID", "user_id", id, "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user", h.logger)
		}
		return
	}

	// Validate phone number
	if req.Phone == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number is required", h.logger)
		return
	}

	// Check if the phone number is already in use by another user
	existingUser, err := h.db.GetUserByPhone(r.Context(), req.Phone)
	if err == nil && existingUser.UserID != id {
		// Phone number is already used by another user
		RespondWithError(w, http.StatusBadRequest, "Phone number already in use by another user", h.logger)
		return
	} else if err != nil && err != sql.ErrNoRows {
		// Database error
		h.logger.ErrorContext(r.Context(), "Failed to check for existing user", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user", h.logger)
		return
	}

	// For now, let's just respond with the user data and a message
	// In a real implementation, you'd update the user in the database
	// We fetch the existing user to get CreatedAt, as it's not part of updateUserRequest
	existingDbUser, err := h.db.GetUserByID(r.Context(), id)
	if err != nil {
		// This should ideally not happen if the previous check passed, but good for safety
		h.logger.ErrorContext(r.Context(), "Failed to get existing user for update response", "user_id", id, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to finalize user update", h.logger)
		return
	}

	var namePtr *string
	if req.Name != "" {
		namePtr = &req.Name
	}
	
	var createdAtStr string
	if existingDbUser.CreatedAt.Valid {
		createdAtStr = existingDbUser.CreatedAt.Time.Format(time.RFC3339)
	}

	apiUser := UserAPIResponse{
		ID:    id,
		Phone: req.Phone,
		Name:  namePtr,
		CreatedAt: createdAtStr, // Preserve original creation timestamp
		Role:  existingDbUser.Role, // Include role from fetched user
	}

	// TODO: Implement actual user update logic in the database
	// For now, just pretend we updated the user
	// TODO: Implement actual update logic for user details including role
	// Placeholder: fetch the user again to return current data including potentially unchanged role
	dbUpdatedUser, _ := h.db.GetUserByID(r.Context(), id) // Ignoring error for placeholder
	var updatedNamePtr *string
	if dbUpdatedUser.Name.Valid {
		updatedNamePtr = &dbUpdatedUser.Name.String
	}
	var updatedCreatedAtStr string
	if dbUpdatedUser.CreatedAt.Valid {
		updatedCreatedAtStr = dbUpdatedUser.CreatedAt.Time.Format(time.RFC3339)
	}
	apiUser = UserAPIResponse{
		ID:        dbUpdatedUser.UserID,
		Phone:     dbUpdatedUser.Phone,
		Name:      updatedNamePtr,
		CreatedAt: updatedCreatedAtStr,
		Role:      dbUpdatedUser.Role, // Include role from fetched user
	}

	RespondWithJSON(w, http.StatusOK, apiUser, h.logger)
}

// AdminDeleteUser handles DELETE /api/admin/users/{id}
// @Summary Delete a user (Admin)
// @Description Deletes a user by their ID. Requires admin authentication.
// @Tags admin/users
// @Produce json
// @Param id path int64 true "User ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/users/{id} [delete]
func (h *AdminUserHandler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger, "error", err)
		return
	}

	// Optional: Check if user exists before attempting delete, to return 404 if not found.
	// However, DELETE is often idempotent, so an error from db.DeleteUser if not found might be okay too.
	// For a better UX, checking first is good.
	_, err = h.db.GetUserByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
		} else {
			h.logger.ErrorContext(r.Context(), "Failed to check user before delete", "user_id", id, "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete user", h.logger)
		}
		return
	}

	err = h.db.DeleteUser(r.Context(), id)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to delete user", "user_id", id, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete user", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"}, h.logger)
}

// Helper function to validate role
func isValidRole(role string) bool {
	switch role {
	case "admin", "owl", "guest":
		return true
	default:
		return false
	}
}

// Helper function to dereference string pointer or return empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
} 