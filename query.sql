-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY id ASC;

-- name: CreateUser :one
INSERT INTO users (name) VALUES (?) RETURNING id, name;

-- name: UpdateUser :exec
UPDATE users set name = ? WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
