-- name: GetEmergencyContacts :many
SELECT contact_id, name, number, description, is_default, is_active, display_order, created_at, updated_at
FROM emergency_contacts 
WHERE is_active = 1
ORDER BY display_order ASC, name ASC;

-- name: GetEmergencyContactByID :one
SELECT contact_id, name, number, description, is_default, is_active, display_order, created_at, updated_at
FROM emergency_contacts 
WHERE contact_id = ? AND is_active = 1;

-- name: CreateEmergencyContact :one
INSERT INTO emergency_contacts (name, number, description, is_default, display_order)
VALUES (?, ?, ?, ?, ?)
RETURNING contact_id, name, number, description, is_default, is_active, display_order, created_at, updated_at;

-- name: UpdateEmergencyContact :one
UPDATE emergency_contacts 
SET name = ?, number = ?, description = ?, is_default = ?, display_order = ?, updated_at = CURRENT_TIMESTAMP
WHERE contact_id = ? AND is_active = 1
RETURNING contact_id, name, number, description, is_default, is_active, display_order, created_at, updated_at;

-- name: DeleteEmergencyContact :exec
UPDATE emergency_contacts 
SET is_active = 0, updated_at = CURRENT_TIMESTAMP
WHERE contact_id = ?;

-- name: SetDefaultEmergencyContact :exec
UPDATE emergency_contacts 
SET is_default = CASE WHEN contact_id = ? THEN 1 ELSE 0 END,
    updated_at = CURRENT_TIMESTAMP
WHERE is_active = 1;

-- name: GetDefaultEmergencyContact :one
SELECT contact_id, name, number, description, is_default, is_active, display_order, created_at, updated_at
FROM emergency_contacts 
WHERE is_default = 1 AND is_active = 1
LIMIT 1; 