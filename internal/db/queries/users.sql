-- name: CreateUser :one
INSERT INTO users (
    phone,
    name,
    role
) VALUES (
    ?,
    ?,
    COALESCE(sqlc.narg('role'), 'guest') -- Use guest if role is not provided
)
RETURNING user_id, phone, name, created_at, role;

-- name: GetUserByPhone :one
SELECT user_id, phone, name, created_at, role FROM users
WHERE phone = ?;

-- name: GetUserByID :one
SELECT user_id, phone, name, created_at, role FROM users
WHERE user_id = ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: ListUsers :many
SELECT user_id, phone, name, created_at, role FROM users
WHERE (sqlc.narg('search_term') IS NULL OR name LIKE sqlc.narg('search_term') OR phone LIKE sqlc.narg('search_term'))
ORDER BY name;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?;

-- name: AdminBulkDeleteUsers :exec
DELETE FROM users
WHERE user_id IN (sqlc.slice('user_ids'));

-- name: UpdateUser :one
UPDATE users
SET
    phone = COALESCE(sqlc.narg('phone'), phone),
    name = COALESCE(sqlc.narg('name'), name),
    role = COALESCE(sqlc.narg('role'), role)
WHERE
    user_id = sqlc.arg('user_id')
RETURNING user_id, phone, name, created_at, role; 