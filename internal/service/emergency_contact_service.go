package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	db "night-owls-go/internal/db/sqlc_generated"
)

var (
	ErrEmergencyContactNotFound = errors.New("emergency contact not found")
	ErrInvalidContactData       = errors.New("invalid contact data")
	ErrCannotDeleteDefault      = errors.New("cannot delete the default emergency contact")
)

type EmergencyContactService struct {
	querier db.Querier
	logger  *slog.Logger
}

func NewEmergencyContactService(querier db.Querier, logger *slog.Logger) *EmergencyContactService {
	return &EmergencyContactService{
		querier: querier,
		logger:  logger.With("service", "EmergencyContactService"),
	}
}

// GetEmergencyContacts returns all active emergency contacts ordered by display order
func (s *EmergencyContactService) GetEmergencyContacts(ctx context.Context) ([]db.EmergencyContact, error) {
	contacts, err := s.querier.GetEmergencyContacts(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get emergency contacts", "error", err)
		return nil, err
	}
	return contacts, nil
}

// GetEmergencyContactByID returns a specific emergency contact by ID
func (s *EmergencyContactService) GetEmergencyContactByID(ctx context.Context, contactID int64) (db.EmergencyContact, error) {
	contact, err := s.querier.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get emergency contact by ID", "contact_id", contactID, "error", err)
		return db.EmergencyContact{}, ErrEmergencyContactNotFound
	}
	return contact, nil
}

// GetDefaultEmergencyContact returns the default emergency contact (usually RUSA)
func (s *EmergencyContactService) GetDefaultEmergencyContact(ctx context.Context) (db.EmergencyContact, error) {
	contact, err := s.querier.GetDefaultEmergencyContact(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get default emergency contact", "error", err)
		return db.EmergencyContact{}, ErrEmergencyContactNotFound
	}
	return contact, nil
}

// CreateEmergencyContact creates a new emergency contact
func (s *EmergencyContactService) CreateEmergencyContact(ctx context.Context, name, number, description string, isDefault bool, displayOrder int64) (db.EmergencyContact, error) {
	if name == "" || number == "" {
		return db.EmergencyContact{}, ErrInvalidContactData
	}

	// If this is being set as default, update all others to not be default
	if isDefault {
		err := s.querier.SetDefaultEmergencyContact(ctx, 0) // This will set all to false first
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to clear default emergency contacts", "error", err)
			return db.EmergencyContact{}, err
		}
	}

	contact, err := s.querier.CreateEmergencyContact(ctx, db.CreateEmergencyContactParams{
		Name:         name,
		Number:       number,
		Description:  sql.NullString{String: description, Valid: description != ""},
		IsDefault:    isDefault,
		DisplayOrder: displayOrder,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create emergency contact", "name", name, "error", err)
		return db.EmergencyContact{}, err
	}

	// If this was set as default, now set it properly
	if isDefault {
		err = s.querier.SetDefaultEmergencyContact(ctx, contact.ContactID)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to set new contact as default", "contact_id", contact.ContactID, "error", err)
			return db.EmergencyContact{}, err
		}
		// Refresh the contact to get updated default status
		contact.IsDefault = true
	}

	s.logger.InfoContext(ctx, "Emergency contact created successfully", "contact_id", contact.ContactID, "name", name)
	return contact, nil
}

// UpdateEmergencyContact updates an existing emergency contact
func (s *EmergencyContactService) UpdateEmergencyContact(ctx context.Context, contactID int64, name, number, description string, isDefault bool, displayOrder int64) (db.EmergencyContact, error) {
	if name == "" || number == "" {
		return db.EmergencyContact{}, ErrInvalidContactData
	}

	// Check if contact exists
	_, err := s.querier.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		return db.EmergencyContact{}, ErrEmergencyContactNotFound
	}

	// If this is being set as default, update all others to not be default
	if isDefault {
		err := s.querier.SetDefaultEmergencyContact(ctx, contactID)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to set default emergency contact", "contact_id", contactID, "error", err)
			return db.EmergencyContact{}, err
		}
	}

	contact, err := s.querier.UpdateEmergencyContact(ctx, db.UpdateEmergencyContactParams{
		ContactID:    contactID,
		Name:         name,
		Number:       number,
		Description:  sql.NullString{String: description, Valid: description != ""},
		IsDefault:    isDefault,
		DisplayOrder: displayOrder,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to update emergency contact", "contact_id", contactID, "error", err)
		return db.EmergencyContact{}, err
	}

	s.logger.InfoContext(ctx, "Emergency contact updated successfully", "contact_id", contactID, "name", name)
	return contact, nil
}

// DeleteEmergencyContact soft deletes an emergency contact
func (s *EmergencyContactService) DeleteEmergencyContact(ctx context.Context, contactID int64) error {
	// Check if contact exists and if it's the default
	contact, err := s.querier.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		return ErrEmergencyContactNotFound
	}

	if contact.IsDefault {
		return ErrCannotDeleteDefault
	}

	err = s.querier.DeleteEmergencyContact(ctx, contactID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete emergency contact", "contact_id", contactID, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "Emergency contact deleted successfully", "contact_id", contactID)
	return nil
}

// SetDefaultEmergencyContact sets a specific contact as the default
func (s *EmergencyContactService) SetDefaultEmergencyContact(ctx context.Context, contactID int64) error {
	// Check if contact exists
	_, err := s.querier.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		return ErrEmergencyContactNotFound
	}

	err = s.querier.SetDefaultEmergencyContact(ctx, contactID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to set default emergency contact", "contact_id", contactID, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "Default emergency contact updated", "contact_id", contactID)
	return nil
} 