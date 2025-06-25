-- name: CreateTokenForUser :one
INSERT INTO password_resets (user_id, token_hash, expires_at)
VALUES ($1, $2, $3) ON CONFLICT (user_id) DO
    UPDATE
        SET token_hash = $2, expires_at = $3
RETURNING id;

-- name: GetUserForToken :one
SELECT
    password_resets.id AS reset_token_id,
    password_resets.expires_at,
    users.id AS user_id,
    users.name,
    users.email,
    users.password_hash,
    users.is_admin
FROM password_resets
    JOIN users on users.id = password_resets.user_id
WHERE password_resets.token_hash = $1;

-- name: DeleteTokenById :exec
DELETE FROM password_resets
WHERE id = $1;
