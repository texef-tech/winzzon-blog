package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const getSEOByPostID = `
SELECT id, post_id, meta_title, meta_description, og_title, og_description, og_image, canonical_url, focus_keyword, schema_type, no_index, created_at, updated_at
FROM seo_meta WHERE post_id = $1
`

func (q *Queries) GetSEOByPostID(ctx context.Context, postID uuid.UUID) (SeoMeta, error) {
	row := q.db.QueryRow(ctx, getSEOByPostID, postID)
	var s SeoMeta
	err := row.Scan(&s.ID, &s.PostID, &s.MetaTitle, &s.MetaDescription, &s.OgTitle, &s.OgDescription, &s.OgImage, &s.CanonicalURL, &s.FocusKeyword, &s.SchemaType, &s.NoIndex, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

const upsertSEO = `
INSERT INTO seo_meta (post_id, meta_title, meta_description, og_title, og_description, og_image, canonical_url, focus_keyword, schema_type, no_index)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (post_id) DO UPDATE SET
    meta_title = EXCLUDED.meta_title,
    meta_description = EXCLUDED.meta_description,
    og_title = EXCLUDED.og_title,
    og_description = EXCLUDED.og_description,
    og_image = EXCLUDED.og_image,
    canonical_url = EXCLUDED.canonical_url,
    focus_keyword = EXCLUDED.focus_keyword,
    schema_type = EXCLUDED.schema_type,
    no_index = EXCLUDED.no_index,
    updated_at = NOW()
RETURNING id, post_id, meta_title, meta_description, og_title, og_description, og_image, canonical_url, focus_keyword, schema_type, no_index, created_at, updated_at
`

type UpsertSEOParams struct {
	PostID          uuid.UUID
	MetaTitle       pgtype.Text
	MetaDescription pgtype.Text
	OgTitle         pgtype.Text
	OgDescription   pgtype.Text
	OgImage         pgtype.Text
	CanonicalURL    pgtype.Text
	FocusKeyword    pgtype.Text
	SchemaType      pgtype.Text
	NoIndex         pgtype.Bool
}

func (q *Queries) UpsertSEO(ctx context.Context, arg UpsertSEOParams) (SeoMeta, error) {
	row := q.db.QueryRow(ctx, upsertSEO,
		arg.PostID, arg.MetaTitle, arg.MetaDescription, arg.OgTitle, arg.OgDescription,
		arg.OgImage, arg.CanonicalURL, arg.FocusKeyword, arg.SchemaType, arg.NoIndex,
	)
	var s SeoMeta
	err := row.Scan(&s.ID, &s.PostID, &s.MetaTitle, &s.MetaDescription, &s.OgTitle, &s.OgDescription, &s.OgImage, &s.CanonicalURL, &s.FocusKeyword, &s.SchemaType, &s.NoIndex, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}
