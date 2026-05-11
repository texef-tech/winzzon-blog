package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createTag = `INSERT INTO tags (name, slug) VALUES ($1, $2) RETURNING id, name, slug, created_at`

type CreateTagParams struct {
	Name string
	Slug string
}

func (q *Queries) CreateTag(ctx context.Context, arg CreateTagParams) (Tag, error) {
	row := q.db.QueryRow(ctx, createTag, arg.Name, arg.Slug)
	var t Tag
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	return t, err
}

const getTagByID = `SELECT id, name, slug, created_at FROM tags WHERE id = $1`

func (q *Queries) GetTagByID(ctx context.Context, id uuid.UUID) (Tag, error) {
	row := q.db.QueryRow(ctx, getTagByID, id)
	var t Tag
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	return t, err
}

const getTagBySlug = `SELECT id, name, slug, created_at FROM tags WHERE slug = $1`

func (q *Queries) GetTagBySlug(ctx context.Context, slug string) (Tag, error) {
	row := q.db.QueryRow(ctx, getTagBySlug, slug)
	var t Tag
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	return t, err
}

const listTags = `SELECT id, name, slug, created_at FROM tags ORDER BY name`

func (q *Queries) ListTags(ctx context.Context) ([]Tag, error) {
	rows, err := q.db.Query(ctx, listTags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

const updateTag = `UPDATE tags SET name = $2, slug = $3 WHERE id = $1 RETURNING id, name, slug, created_at`

type UpdateTagParams struct {
	ID   uuid.UUID
	Name string
	Slug string
}

func (q *Queries) UpdateTag(ctx context.Context, arg UpdateTagParams) (Tag, error) {
	row := q.db.QueryRow(ctx, updateTag, arg.ID, arg.Name, arg.Slug)
	var t Tag
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	return t, err
}

const deleteTag = `DELETE FROM tags WHERE id = $1`

func (q *Queries) DeleteTag(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteTag, id)
	return err
}

const addPostTag = `INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`

func (q *Queries) AddPostTag(ctx context.Context, postID uuid.UUID, tagID uuid.UUID) error {
	_, err := q.db.Exec(ctx, addPostTag, postID, tagID)
	return err
}

const removePostTags = `DELETE FROM post_tags WHERE post_id = $1`

func (q *Queries) RemovePostTags(ctx context.Context, postID uuid.UUID) error {
	_, err := q.db.Exec(ctx, removePostTags, postID)
	return err
}

const listTagsForPost = `
SELECT t.id, t.name, t.slug, t.created_at FROM tags t
JOIN post_tags pt ON pt.tag_id = t.id
WHERE pt.post_id = $1 ORDER BY t.name
`

func (q *Queries) ListTagsForPost(ctx context.Context, postID uuid.UUID) ([]Tag, error) {
	rows, err := q.db.Query(ctx, listTagsForPost, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}
