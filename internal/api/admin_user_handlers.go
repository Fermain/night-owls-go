package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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
	Phone string `json:"phone"`
	Name  string `json:"name"`
	Role  string `json:"role"` // Make role required since frontend always sends it
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
	db     db.Querier
	logger *slog.Logger
}

// NewAdminUserHandler creates a new AdminUserHandler.
func NewAdminUserHandler(db db.Querier, logger *slog.Logger) *AdminUserHandler {
	return &AdminUserHandler{
		db:     db,
		logger: logger.With("handler", "AdminUserHandler"),
	}
}

// AdminListUsers handles GET /api/admin/users
// @Summary List all users (Admin)
// @Description Get a list of all users in the system. Requires admin authentication.
// @Tags admin/users
// @Produce json
// @Param search query string false "Search term to filter users by name or phone"
// @Success 200 {array} UserAPIResponse "List of users"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/users [get]
func (h *AdminUserHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")

	var searchQuery sql.NullString
	if searchTerm != "" {
		searchQuery = sql.NullString{String: "%" + searchTerm + "%", Valid: true}
	} else {
		searchQuery = sql.NullString{Valid: false}
	}

	dbUsers, err := h.db.ListUsers(r.Context(), searchQuery)
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
	// Try multiple methods to extract the ID parameter
	idStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "AdminGetUser called", "id_param", idStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if idStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "users" {
			idStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", idStr)
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse user ID", "id_param", idStr, "error", err)
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
	// Try multiple methods to extract the ID parameter
	idStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "AdminUpdateUser called", "id_param", idStr, "url", r.URL.Path)

	// Alternative method 1: Parse from URL path directly
	if idStr == "" {
		// Extract ID from path manually: /api/admin/users/{id}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		h.logger.InfoContext(r.Context(), "URL path parts", "parts", pathParts)

		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "users" {
			idStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", idStr)
		}
	}

	// Alternative method 2: Check request context for route values
	if idStr == "" {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			h.logger.InfoContext(r.Context(), "Chi route context", "url_params", rctx.URLParams)
			for i, param := range rctx.URLParams.Keys {
				if param == "id" && i < len(rctx.URLParams.Values) {
					idStr = rctx.URLParams.Values[i]
					h.logger.InfoContext(r.Context(), "Found ID in route context", "id_param", idStr)
					break
				}
			}
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse user ID", "id_param", idStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger, "error", err)
		return
	}

	h.logger.InfoContext(r.Context(), "Parsed user ID successfully", "user_id", id)

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

	// Validate role if provided
	if req.Role == "" {
		RespondWithError(w, http.StatusBadRequest, "Role is required", h.logger)
		return
	}

	if !isValidRole(req.Role) {
		RespondWithError(w, http.StatusBadRequest, "Invalid role specified. Must be one of: admin, owl, guest", h.logger)
		return
	}

	// Prepare parameters for database update
	updateParams := db.UpdateUserParams{
		UserID: id,
		Phone:  sql.NullString{String: req.Phone, Valid: req.Phone != ""}, // Assume phone is always provided from frontend form
		Name:   sql.NullString{String: req.Name, Valid: req.Name != ""},   // Will be null if name is empty string
		Role:   sql.NullString{String: req.Role, Valid: true},             // Role is always provided
	}

	// Perform the update
	updatedDbUser, err := h.db.UpdateUser(r.Context(), updateParams)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to update user in database", "user_id", id, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user", h.logger)
		return
	}

	var updatedNamePtr *string
	if updatedDbUser.Name.Valid {
		updatedNamePtr = &updatedDbUser.Name.String
	}
	var updatedCreatedAtStr string
	if updatedDbUser.CreatedAt.Valid {
		updatedCreatedAtStr = updatedDbUser.CreatedAt.Time.Format(time.RFC3339)
	}

	apiUser := UserAPIResponse{
		ID:        updatedDbUser.UserID,
		Phone:     updatedDbUser.Phone,
		Name:      updatedNamePtr,
		CreatedAt: updatedCreatedAtStr,
		Role:      updatedDbUser.Role,
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
	// Try multiple methods to extract the ID parameter
	idStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "AdminDeleteUser called", "id_param", idStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if idStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "users" {
			idStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", idStr)
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse user ID", "id_param", idStr, "error", err)
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

// AdminBulkDeleteUsers handles requests to bulk delete users.
// @Summary Bulk delete users (Admin)
// @Description Delete multiple users by their IDs. Requires admin authentication.
// @Tags admin/users
// @Accept json
// @Produce json
// @Param request body object{user_ids=[]int64} true "List of user IDs to delete"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} ErrorResponse "Invalid request or no user IDs provided"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/users/bulk-delete [post]
func (h *AdminUserHandler) AdminBulkDeleteUsers(w http.ResponseWriter, r *http.Request) {
	type BulkDeleteRequest struct {
		UserIDs []int64 `json:"user_ids"`
	}
	var req BulkDeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err)
		return
	}

	if len(req.UserIDs) == 0 {
		RespondWithError(w, http.StatusBadRequest, "No user IDs provided for deletion", h.logger)
		return
	}

	// Optional: Add a reasonable limit to prevent deletion of too many users at once
	if len(req.UserIDs) > 100 {
		RespondWithError(w, http.StatusBadRequest, "Too many users selected for bulk deletion (max 100)", h.logger)
		return
	}

	err := h.db.AdminBulkDeleteUsers(r.Context(), req.UserIDs)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Error bulk deleting users", "error", err, "user_ids", req.UserIDs)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete users", h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Successfully bulk deleted users", "count", len(req.UserIDs), "user_ids", req.UserIDs)
	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Successfully deleted %d users", len(req.UserIDs)),
	}, h.logger)
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
