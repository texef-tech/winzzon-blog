-- name: CreateCategory :one
INSERT INTO categories (name, slug, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- name: ListCategories :many
SELECT * FROM categories ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories SET name = $2, slug = $3, description = $4
WHERE id = $1
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: AddPostCategory :exec
INSERT INTO post_categories (post_id, category_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePostCategories :exec
DELETE FROM post_categories WHERE post_id = $1;

-- name: ListCategoriesForPost :many
SELECT c.* FROM categories c
JOIN post_categories pc ON pc.category_id = c.id
WHERE pc.post_id = $1
ORDER BY c.name;
