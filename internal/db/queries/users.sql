-- name: CreateUser :one
INSERT INTO users (
    phone,
    name
) VALUES (
    ?,
    ?
)
RETURNING *;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = ?;

-- name: GetUserByID :one
SELECT * FROM users
WHERE user_id = ?;

-- name: ListUsers :many
SELECT * FROM users;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?; 