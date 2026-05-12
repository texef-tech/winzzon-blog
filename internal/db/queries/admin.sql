-- name: GetAdminByUsername :one
SELECT * FROM admin_credentials WHERE username = $1;

-- name: CreateAdmin :one
INSERT INTO admin_credentials (username, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE admin_credentials SET last_login_at = NOW() WHERE id = $1;

-- name: AdminExists :one
SELECT EXISTS(SELECT 1 FROM admin_credentials WHERE username = $1);
