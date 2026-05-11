CREATE TABLE seo_meta (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id          UUID NOT NULL UNIQUE REFERENCES posts(id) ON DELETE CASCADE,
    meta_title       TEXT,
    meta_description TEXT,
    og_title         TEXT,
    og_description   TEXT,
    og_image         TEXT,
    canonical_url    TEXT,
    focus_keyword    TEXT,
    schema_type      TEXT DEFAULT 'Article',
    no_index         BOOLEAN DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
