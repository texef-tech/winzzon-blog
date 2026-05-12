package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createCategory = `
INSERT INTO categories (name, slug, description) VALUES ($1, $2, $3)
RETURNING id, name, slug, description, created_at
`

type CreateCategoryParams struct {
	Name        string
	Slug        string
	Description pgtype.Text
}

func (q *Queries) CreateCategory(ctx context.Context, arg CreateCategoryParams) (Category, error) {
	row := q.db.QueryRow(ctx, createCategory, arg.Name, arg.Slug, arg.Description)
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt)
	return c, err
}

const getCategoryByID = `SELECT id, name, slug, description, created_at FROM categories WHERE id = $1`

func (q *Queries) GetCategoryByID(ctx context.Context, id uuid.UUID) (Category, error) {
	row := q.db.QueryRow(ctx, getCategoryByID, id)
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt)
	return c, err
}

const getCategoryBySlug = `SELECT id, name, slug, description, created_at FROM categories WHERE slug = $1`

func (q *Queries) GetCategoryBySlug(ctx context.Context, slug string) (Category, error) {
	row := q.db.QueryRow(ctx, getCategoryBySlug, slug)
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt)
	return c, err
}

const listCategories = `SELECT id, name, slug, description, created_at FROM categories ORDER BY name`

func (q *Queries) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := q.db.Query(ctx, listCategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

const updateCategory = `
UPDATE categories SET name = $2, slug = $3, description = $4 WHERE id = $1
RETURNING id, name, slug, description, created_at
`

type UpdateCategoryParams struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description pgtype.Text
}

func (q *Queries) UpdateCategory(ctx context.Context, arg UpdateCategoryParams) (Category, error) {
	row := q.db.QueryRow(ctx, updateCategory, arg.ID, arg.Name, arg.Slug, arg.Description)
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt)
	return c, err
}

const deleteCategory = `DELETE FROM categories WHERE id = $1`

func (q *Queries) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteCategory, id)
	return err
}

const addPostCategory = `INSERT INTO post_categories (post_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`

func (q *Queries) AddPostCategory(ctx context.Context, postID uuid.UUID, categoryID uuid.UUID) error {
	_, err := q.db.Exec(ctx, addPostCategory, postID, categoryID)
	return err
}

const removePostCategories = `DELETE FROM post_categories WHERE post_id = $1`

func (q *Queries) RemovePostCategories(ctx context.Context, postID uuid.UUID) error {
	_, err := q.db.Exec(ctx, removePostCategories, postID)
	return err
}

const listCategoriesForPost = `
SELECT c.id, c.name, c.slug, c.description, c.created_at FROM categories c
JOIN post_categories pc ON pc.category_id = c.id
WHERE pc.post_id = $1 ORDER BY c.name
`

func (q *Queries) ListCategoriesForPost(ctx context.Context, postID uuid.UUID) ([]Category, error) {
	rows, err := q.db.Query(ctx, listCategoriesForPost, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}
