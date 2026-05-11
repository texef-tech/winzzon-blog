package sqlc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createPost = `
INSERT INTO posts (title, slug, body, body_plain, excerpt, reading_time, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
`

type CreatePostParams struct {
	Title       string
	Slug        string
	Body        string
	BodyPlain   pgtype.Text
	Excerpt     pgtype.Text
	ReadingTime pgtype.Int4
	Status      string
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	row := q.db.QueryRow(ctx, createPost,
		arg.Title, arg.Slug, arg.Body, arg.BodyPlain, arg.Excerpt, arg.ReadingTime, arg.Status,
	)
	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

const getPostByID = `
SELECT id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
FROM posts WHERE id = $1 AND deleted_at IS NULL
`

func (q *Queries) GetPostByID(ctx context.Context, id uuid.UUID) (Post, error) {
	row := q.db.QueryRow(ctx, getPostByID, id)
	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

const getPostBySlug = `
SELECT id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
FROM posts WHERE slug = $1 AND status = 'published' AND deleted_at IS NULL
`

func (q *Queries) GetPostBySlug(ctx context.Context, slug string) (Post, error) {
	row := q.db.QueryRow(ctx, getPostBySlug, slug)
	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

const listPublishedPosts = `
SELECT id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
FROM posts WHERE status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC LIMIT $1 OFFSET $2
`

func (q *Queries) ListPublishedPosts(ctx context.Context, limit int32, offset int32) ([]Post, error) {
	rows, err := q.db.Query(ctx, listPublishedPosts, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

const countPublishedPosts = `SELECT COUNT(*) FROM posts WHERE status = 'published' AND deleted_at IS NULL`

func (q *Queries) CountPublishedPosts(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countPublishedPosts)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listAllPosts = `
SELECT id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
FROM posts WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2
`

func (q *Queries) ListAllPosts(ctx context.Context, limit int32, offset int32) ([]Post, error) {
	rows, err := q.db.Query(ctx, listAllPosts, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

const countAllPosts = `SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL`

func (q *Queries) CountAllPosts(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countAllPosts)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const updatePost = `
UPDATE posts SET title = $2, body = $3, body_plain = $4, excerpt = $5, reading_time = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
`

type UpdatePostParams struct {
	ID          uuid.UUID
	Title       string
	Body        string
	BodyPlain   pgtype.Text
	Excerpt     pgtype.Text
	ReadingTime pgtype.Int4
}

func (q *Queries) UpdatePost(ctx context.Context, arg UpdatePostParams) (Post, error) {
	row := q.db.QueryRow(ctx, updatePost, arg.ID, arg.Title, arg.Body, arg.BodyPlain, arg.Excerpt, arg.ReadingTime)
	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

const publishPost = `
UPDATE posts SET status = 'published', published_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
`

func (q *Queries) PublishPost(ctx context.Context, id uuid.UUID) (Post, error) {
	row := q.db.QueryRow(ctx, publishPost, id)
	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

const unpublishPost = `
UPDATE posts SET status = 'draft', updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
`

func (q *Queries) UnpublishPost(ctx context.Context, id uuid.UUID) (Post, error) {
	row := q.db.QueryRow(ctx, unpublishPost, id)
	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

const softDeletePost = `UPDATE posts SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

func (q *Queries) SoftDeletePost(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, softDeletePost, id)
	return err
}

const slugExists = `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1)`

func (q *Queries) SlugExists(ctx context.Context, slug string) (bool, error) {
	row := q.db.QueryRow(ctx, slugExists, slug)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const listPublishedPostsByCategory = `
SELECT p.id, p.title, p.slug, p.body, p.body_plain, p.excerpt, p.reading_time, p.status, p.published_at, p.deleted_at, p.created_at, p.updated_at
FROM posts p
JOIN post_categories pc ON pc.post_id = p.id
JOIN categories c ON c.id = pc.category_id
WHERE c.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL
ORDER BY p.published_at DESC LIMIT $2 OFFSET $3
`

func (q *Queries) ListPublishedPostsByCategory(ctx context.Context, categorySlug string, limit int32, offset int32) ([]Post, error) {
	rows, err := q.db.Query(ctx, listPublishedPostsByCategory, categorySlug, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

const countPublishedPostsByCategory = `
SELECT COUNT(*) FROM posts p
JOIN post_categories pc ON pc.post_id = p.id
JOIN categories c ON c.id = pc.category_id
WHERE c.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL
`

func (q *Queries) CountPublishedPostsByCategory(ctx context.Context, categorySlug string) (int64, error) {
	row := q.db.QueryRow(ctx, countPublishedPostsByCategory, categorySlug)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listPublishedPostsByTag = `
SELECT p.id, p.title, p.slug, p.body, p.body_plain, p.excerpt, p.reading_time, p.status, p.published_at, p.deleted_at, p.created_at, p.updated_at
FROM posts p
JOIN post_tags pt ON pt.post_id = p.id
JOIN tags t ON t.id = pt.tag_id
WHERE t.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL
ORDER BY p.published_at DESC LIMIT $2 OFFSET $3
`

func (q *Queries) ListPublishedPostsByTag(ctx context.Context, tagSlug string, limit int32, offset int32) ([]Post, error) {
	rows, err := q.db.Query(ctx, listPublishedPostsByTag, tagSlug, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

const countPublishedPostsByTag = `
SELECT COUNT(*) FROM posts p
JOIN post_tags pt ON pt.post_id = p.id
JOIN tags t ON t.id = pt.tag_id
WHERE t.slug = $1 AND p.status = 'published' AND p.deleted_at IS NULL
`

func (q *Queries) CountPublishedPostsByTag(ctx context.Context, tagSlug string) (int64, error) {
	row := q.db.QueryRow(ctx, countPublishedPostsByTag, tagSlug)
	var count int64
	err := row.Scan(&count)
	return count, err
}

type SitemapPost struct {
	ID        uuid.UUID
	Slug      string
	UpdatedAt time.Time
}

const listAllPublishedPostsForSitemap = `
SELECT id, slug, updated_at FROM posts
WHERE status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC
`

func (q *Queries) ListAllPublishedPostsForSitemap(ctx context.Context) ([]SitemapPost, error) {
	rows, err := q.db.Query(ctx, listAllPublishedPostsForSitemap)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SitemapPost
	for rows.Next() {
		var p SitemapPost
		if err := rows.Scan(&p.ID, &p.Slug, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

const listRecentPublishedPosts = `
SELECT id, title, slug, body, body_plain, excerpt, reading_time, status, published_at, deleted_at, created_at, updated_at
FROM posts WHERE status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC LIMIT $1
`

func (q *Queries) ListRecentPublishedPosts(ctx context.Context, limit int32) ([]Post, error) {
	rows, err := q.db.Query(ctx, listRecentPublishedPosts, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.BodyPlain, &p.Excerpt, &p.ReadingTime, &p.Status, &p.PublishedAt, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
