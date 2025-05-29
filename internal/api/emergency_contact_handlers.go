package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
)

type EmergencyContactHandler struct {
	emergencyContactService *service.EmergencyContactService
	logger                  *slog.Logger
}

func NewEmergencyContactHandler(emergencyContactService *service.EmergencyContactService, logger *slog.Logger) *EmergencyContactHandler {
	return &EmergencyContactHandler{
		emergencyContactService: emergencyContactService,
		logger:                  logger.With("handler", "EmergencyContactHandler"),
	}
}

// EmergencyContactResponse represents the API response format
type EmergencyContactResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Number       string `json:"number"`
	Description  string `json:"description"`
	IsDefault    bool   `json:"is_default"`
	DisplayOrder int64  `json:"display_order"`
}

// CreateEmergencyContactRequest represents the request to create an emergency contact
type CreateEmergencyContactRequest struct {
	Name         string `json:"name"`
	Number       string `json:"number"`
	Description  string `json:"description"`
	IsDefault    bool   `json:"is_default"`
	DisplayOrder int64  `json:"display_order"`
}

// UpdateEmergencyContactRequest represents the request to update an emergency contact
type UpdateEmergencyContactRequest struct {
	Name         string `json:"name"`
	Number       string `json:"number"`
	Description  string `json:"description"`
	IsDefault    bool   `json:"is_default"`
	DisplayOrder int64  `json:"display_order"`
}

// GetEmergencyContactsHandler handles GET /api/emergency-contacts
// @Summary Get emergency contacts
// @Description Returns all active emergency contacts for public use
// @Tags emergency-contacts
// @Produce json
// @Success 200 {array} EmergencyContactResponse "List of emergency contacts"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/emergency-contacts [get]
func (h *EmergencyContactHandler) GetEmergencyContactsHandler(w http.ResponseWriter, r *http.Request) {
	contacts, err := h.emergencyContactService.GetEmergencyContacts(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get emergency contacts", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get emergency contacts", h.logger, "error", err.Error())
		return
	}

	response := make([]EmergencyContactResponse, len(contacts))
	for i, contact := range contacts {
		response[i] = EmergencyContactResponse{
			ID:           contact.ContactID,
			Name:         contact.Name,
			Number:       contact.Number,
			Description:  contact.Description.String,
			IsDefault:    contact.IsDefault,
			DisplayOrder: contact.DisplayOrder,
		}
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetDefaultEmergencyContactHandler handles GET /api/emergency-contacts/default
// @Summary Get default emergency contact
// @Description Returns the default emergency contact (usually RUSA)
// @Tags emergency-contacts
// @Produce json
// @Success 200 {object} EmergencyContactResponse "Default emergency contact"
// @Failure 404 {object} ErrorResponse "No default emergency contact found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/emergency-contacts/default [get]
func (h *EmergencyContactHandler) GetDefaultEmergencyContactHandler(w http.ResponseWriter, r *http.Request) {
	contact, err := h.emergencyContactService.GetDefaultEmergencyContact(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrEmergencyContactNotFound) {
			RespondWithError(w, http.StatusNotFound, "No default emergency contact found", h.logger)
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to get default emergency contact", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get default emergency contact", h.logger, "error", err.Error())
		return
	}

	response := EmergencyContactResponse{
		ID:           contact.ContactID,
		Name:         contact.Name,
		Number:       contact.Number,
		Description:  contact.Description.String,
		IsDefault:    contact.IsDefault,
		DisplayOrder: contact.DisplayOrder,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// AdminGetEmergencyContactsHandler handles GET /api/admin/emergency-contacts
// @Summary Admin: Get emergency contacts
// @Description Returns all emergency contacts for admin management
// @Tags admin-emergency-contacts
// @Produce json
// @Success 200 {array} EmergencyContactResponse "List of emergency contacts"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/emergency-contacts [get]
func (h *EmergencyContactHandler) AdminGetEmergencyContactsHandler(w http.ResponseWriter, r *http.Request) {
	h.GetEmergencyContactsHandler(w, r) // Same logic as public endpoint
}

// AdminGetEmergencyContactHandler handles GET /api/admin/emergency-contacts/{id}
// @Summary Admin: Get emergency contact by ID
// @Description Returns a specific emergency contact by ID
// @Tags admin-emergency-contacts
// @Produce json
// @Param id path int true "Emergency Contact ID"
// @Success 200 {object} EmergencyContactResponse "Emergency contact details"
// @Failure 400 {object} ErrorResponse "Invalid contact ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure 404 {object} ErrorResponse "Emergency contact not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/emergency-contacts/{id} [get]
func (h *EmergencyContactHandler) AdminGetEmergencyContactHandler(w http.ResponseWriter, r *http.Request) {
	contactIDStr := chi.URLParam(r, "id")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid contact ID", h.logger, "contact_id", contactIDStr)
		return
	}

	contact, err := h.emergencyContactService.GetEmergencyContactByID(r.Context(), contactID)
	if err != nil {
		if errors.Is(err, service.ErrEmergencyContactNotFound) {
			RespondWithError(w, http.StatusNotFound, "Emergency contact not found", h.logger, "contact_id", contactID)
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to get emergency contact", "contact_id", contactID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get emergency contact", h.logger, "error", err.Error())
		return
	}

	response := EmergencyContactResponse{
		ID:           contact.ContactID,
		Name:         contact.Name,
		Number:       contact.Number,
		Description:  contact.Description.String,
		IsDefault:    contact.IsDefault,
		DisplayOrder: contact.DisplayOrder,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// AdminCreateEmergencyContactHandler handles POST /api/admin/emergency-contacts
// @Summary Admin: Create emergency contact
// @Description Creates a new emergency contact
// @Tags admin-emergency-contacts
// @Accept json
// @Produce json
// @Param request body CreateEmergencyContactRequest true "Emergency contact data"
// @Success 201 {object} EmergencyContactResponse "Created emergency contact"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/emergency-contacts [post]
func (h *EmergencyContactHandler) AdminCreateEmergencyContactHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateEmergencyContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger, "error", err.Error())
		return
	}

	if req.Name == "" || req.Number == "" {
		RespondWithError(w, http.StatusBadRequest, "Name and number are required", h.logger)
		return
	}

	contact, err := h.emergencyContactService.CreateEmergencyContact(
		r.Context(),
		req.Name,
		req.Number,
		req.Description,
		req.IsDefault,
		req.DisplayOrder,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidContactData) {
			RespondWithError(w, http.StatusBadRequest, "Invalid contact data", h.logger, "error", err.Error())
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to create emergency contact", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create emergency contact", h.logger, "error", err.Error())
		return
	}

	response := EmergencyContactResponse{
		ID:           contact.ContactID,
		Name:         contact.Name,
		Number:       contact.Number,
		Description:  contact.Description.String,
		IsDefault:    contact.IsDefault,
		DisplayOrder: contact.DisplayOrder,
	}

	RespondWithJSON(w, http.StatusCreated, response, h.logger)
}

// AdminUpdateEmergencyContactHandler handles PUT /api/admin/emergency-contacts/{id}
// @Summary Admin: Update emergency contact
// @Description Updates an existing emergency contact
// @Tags admin-emergency-contacts
// @Accept json
// @Produce json
// @Param id path int true "Emergency Contact ID"
// @Param request body UpdateEmergencyContactRequest true "Emergency contact data"
// @Success 200 {object} EmergencyContactResponse "Updated emergency contact"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure 404 {object} ErrorResponse "Emergency contact not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/emergency-contacts/{id} [put]
func (h *EmergencyContactHandler) AdminUpdateEmergencyContactHandler(w http.ResponseWriter, r *http.Request) {
	contactIDStr := chi.URLParam(r, "id")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid contact ID", h.logger, "contact_id", contactIDStr)
		return
	}

	var req UpdateEmergencyContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger, "error", err.Error())
		return
	}

	if req.Name == "" || req.Number == "" {
		RespondWithError(w, http.StatusBadRequest, "Name and number are required", h.logger)
		return
	}

	contact, err := h.emergencyContactService.UpdateEmergencyContact(
		r.Context(),
		contactID,
		req.Name,
		req.Number,
		req.Description,
		req.IsDefault,
		req.DisplayOrder,
	)
	if err != nil {
		if errors.Is(err, service.ErrEmergencyContactNotFound) {
			RespondWithError(w, http.StatusNotFound, "Emergency contact not found", h.logger, "contact_id", contactID)
			return
		}
		if errors.Is(err, service.ErrInvalidContactData) {
			RespondWithError(w, http.StatusBadRequest, "Invalid contact data", h.logger, "error", err.Error())
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to update emergency contact", "contact_id", contactID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update emergency contact", h.logger, "error", err.Error())
		return
	}

	response := EmergencyContactResponse{
		ID:           contact.ContactID,
		Name:         contact.Name,
		Number:       contact.Number,
		Description:  contact.Description.String,
		IsDefault:    contact.IsDefault,
		DisplayOrder: contact.DisplayOrder,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// AdminDeleteEmergencyContactHandler handles DELETE /api/admin/emergency-contacts/{id}
// @Summary Admin: Delete emergency contact
// @Description Deletes an emergency contact (cannot delete default contact)
// @Tags admin-emergency-contacts
// @Param id path int true "Emergency Contact ID"
// @Success 204 "Emergency contact deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid contact ID or cannot delete default contact"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure 404 {object} ErrorResponse "Emergency contact not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/emergency-contacts/{id} [delete]
func (h *EmergencyContactHandler) AdminDeleteEmergencyContactHandler(w http.ResponseWriter, r *http.Request) {
	contactIDStr := chi.URLParam(r, "id")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid contact ID", h.logger, "contact_id", contactIDStr)
		return
	}

	err = h.emergencyContactService.DeleteEmergencyContact(r.Context(), contactID)
	if err != nil {
		if errors.Is(err, service.ErrEmergencyContactNotFound) {
			RespondWithError(w, http.StatusNotFound, "Emergency contact not found", h.logger, "contact_id", contactID)
			return
		}
		if errors.Is(err, service.ErrCannotDeleteDefault) {
			RespondWithError(w, http.StatusBadRequest, "Cannot delete the default emergency contact", h.logger, "contact_id", contactID)
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to delete emergency contact", "contact_id", contactID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete emergency contact", h.logger, "error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminSetDefaultEmergencyContactHandler handles PUT /api/admin/emergency-contacts/{id}/default
// @Summary Admin: Set default emergency contact
// @Description Sets a specific emergency contact as the default
// @Tags admin-emergency-contacts
// @Param id path int true "Emergency Contact ID"
// @Success 204 "Default emergency contact updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid contact ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure 404 {object} ErrorResponse "Emergency contact not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/emergency-contacts/{id}/default [put]
func (h *EmergencyContactHandler) AdminSetDefaultEmergencyContactHandler(w http.ResponseWriter, r *http.Request) {
	contactIDStr := chi.URLParam(r, "id")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid contact ID", h.logger, "contact_id", contactIDStr)
		return
	}

	err = h.emergencyContactService.SetDefaultEmergencyContact(r.Context(), contactID)
	if err != nil {
		if errors.Is(err, service.ErrEmergencyContactNotFound) {
			RespondWithError(w, http.StatusNotFound, "Emergency contact not found", h.logger, "contact_id", contactID)
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to set default emergency contact", "contact_id", contactID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to set default emergency contact", h.logger, "error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
