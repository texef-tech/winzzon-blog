-- name: CreateMedia :one
INSERT INTO media (filename, original_name, url, mime_type, size_bytes, alt_text, width, height)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetMediaByID :one
SELECT * FROM media WHERE id = $1;

-- name: ListMedia :many
SELECT * FROM media ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountMedia :one
SELECT COUNT(*) FROM media;

-- name: DeleteMedia :exec
DELETE FROM media WHERE id = $1;
