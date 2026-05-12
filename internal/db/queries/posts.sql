-- name: CreatePost :one
INSERT INTO posts (title, slug, body, body_plain, excerpt, reading_time, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPostBySlug :one
SELECT * FROM posts WHERE slug = $1 AND status = 'published' AND deleted_at IS NULL;

-- name: ListPublishedPosts :many
SELECT * FROM posts
WHERE status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPublishedPosts :one
SELECT COUNT(*) FROM posts WHERE status = 'published' AND deleted_at IS NULL;

-- name: ListAllPosts :many
SELECT * FROM posts
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllPosts :one
SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL;

-- name: UpdatePost :one
UPDATE posts SET
    title = $2,
    body = $3,
    body_plain = $4,
    excerpt = $5,
    reading_time = $6,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: PublishPost :one
UPDATE posts SET status = 'published', published_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UnpublishPost :one
UPDATE posts SET status = 'draft', updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePost :exec
UPDATE posts SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SlugExists :one
SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1);

-- name: ListPublishedPostsByCategory :many
SELECT p.* FROM posts p
JOIN post_categories pc ON pc.post_id = p.id
JOIN categories c ON c.id = pc.category_id
WHERE c.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL
ORDER BY p.published_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPublishedPostsByCategory :one
SELECT COUNT(*) FROM posts p
JOIN post_categories pc ON pc.post_id = p.id
JOIN categories c ON c.id = pc.category_id
WHERE c.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL;

-- name: ListPublishedPostsByTag :many
SELECT p.* FROM posts p
JOIN post_tags pt ON pt.post_id = p.id
JOIN tags t ON t.id = pt.tag_id
WHERE t.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL
ORDER BY p.published_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPublishedPostsByTag :one
SELECT COUNT(*) FROM posts p
JOIN post_tags pt ON pt.post_id = p.id
JOIN tags t ON t.id = pt.tag_id
WHERE t.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL;

-- name: ListAllPublishedPostsForSitemap :many
SELECT id, slug, updated_at FROM posts
WHERE status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC;

-- name: ListRecentPublishedPosts :many
SELECT * FROM posts
WHERE status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC
LIMIT $1;
