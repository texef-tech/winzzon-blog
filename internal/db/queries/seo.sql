-- name: GetSEOByPostID :one
SELECT * FROM seo_meta WHERE post_id = $1;

-- name: UpsertSEO :one
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
RETURNING *;
