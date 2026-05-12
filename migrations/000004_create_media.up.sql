CREATE TABLE media (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename      TEXT NOT NULL,
    original_name TEXT NOT NULL,
    url           TEXT NOT NULL,
    mime_type     TEXT NOT NULL,
    size_bytes    INT NOT NULL,
    alt_text      TEXT,
    width         INT,
    height        INT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
