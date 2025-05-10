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

-- name: ListUsers :many
SELECT user_id, phone, name, created_at, role FROM users;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?; 