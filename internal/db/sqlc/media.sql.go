package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createMedia = `
INSERT INTO media (filename, original_name, url, mime_type, size_bytes, alt_text, width, height)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, filename, original_name, url, mime_type, size_bytes, alt_text, width, height, created_at
`

type CreateMediaParams struct {
	Filename     string
	OriginalName string
	URL          string
	MimeType     string
	SizeBytes    int32
	AltText      pgtype.Text
	Width        pgtype.Int4
	Height       pgtype.Int4
}

func (q *Queries) CreateMedia(ctx context.Context, arg CreateMediaParams) (Media, error) {
	row := q.db.QueryRow(ctx, createMedia,
		arg.Filename, arg.OriginalName, arg.URL, arg.MimeType, arg.SizeBytes, arg.AltText, arg.Width, arg.Height,
	)
	var m Media
	err := row.Scan(&m.ID, &m.Filename, &m.OriginalName, &m.URL, &m.MimeType, &m.SizeBytes, &m.AltText, &m.Width, &m.Height, &m.CreatedAt)
	return m, err
}

const getMediaByID = `SELECT id, filename, original_name, url, mime_type, size_bytes, alt_text, width, height, created_at FROM media WHERE id = $1`

func (q *Queries) GetMediaByID(ctx context.Context, id uuid.UUID) (Media, error) {
	row := q.db.QueryRow(ctx, getMediaByID, id)
	var m Media
	err := row.Scan(&m.ID, &m.Filename, &m.OriginalName, &m.URL, &m.MimeType, &m.SizeBytes, &m.AltText, &m.Width, &m.Height, &m.CreatedAt)
	return m, err
}

const listMedia = `SELECT id, filename, original_name, url, mime_type, size_bytes, alt_text, width, height, created_at FROM media ORDER BY created_at DESC LIMIT $1 OFFSET $2`

func (q *Queries) ListMedia(ctx context.Context, limit int32, offset int32) ([]Media, error) {
	rows, err := q.db.Query(ctx, listMedia, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Media
	for rows.Next() {
		var m Media
		if err := rows.Scan(&m.ID, &m.Filename, &m.OriginalName, &m.URL, &m.MimeType, &m.SizeBytes, &m.AltText, &m.Width, &m.Height, &m.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

const countMedia = `SELECT COUNT(*) FROM media`

func (q *Queries) CountMedia(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countMedia)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteMedia = `DELETE FROM media WHERE id = $1`

func (q *Queries) DeleteMedia(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteMedia, id)
	return err
}
