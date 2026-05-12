package sqlc

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Post struct {
	ID          uuid.UUID
	Title       string
	Slug        string
	Body        string
	BodyPlain   pgtype.Text
	Excerpt     pgtype.Text
	ReadingTime pgtype.Int4
	Status      string
	PublishedAt pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SeoMeta struct {
	ID              uuid.UUID
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
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Category struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description pgtype.Text
	CreatedAt   time.Time
}

type Tag struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
}

type Media struct {
	ID           uuid.UUID
	Filename     string
	OriginalName string
	URL          string
	MimeType     string
	SizeBytes    int32
	AltText      pgtype.Text
	Width        pgtype.Int4
	Height       pgtype.Int4
	CreatedAt    time.Time
}

type AdminCredential struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	LastLoginAt  pgtype.Timestamptz
	CreatedAt    time.Time
}
