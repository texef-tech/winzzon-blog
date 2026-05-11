-- name: CreateTag :one
INSERT INTO tags (name, slug) VALUES ($1, $2) RETURNING *;

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = $1;

-- name: GetTagBySlug :one
SELECT * FROM tags WHERE slug = $1;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: UpdateTag :one
UPDATE tags SET name = $2, slug = $3 WHERE id = $1 RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: AddPostTag :exec
INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePostTags :exec
DELETE FROM post_tags WHERE post_id = $1;

-- name: ListTagsForPost :many
SELECT t.* FROM tags t
JOIN post_tags pt ON pt.tag_id = t.id
WHERE pt.post_id = $1
ORDER BY t.name;
