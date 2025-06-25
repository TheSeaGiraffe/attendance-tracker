-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateNewUser :one
INSERT INTO users (name, email, password_hash, is_admin)
VALUES ($1, $2, $3, $4) RETURNING id;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;
