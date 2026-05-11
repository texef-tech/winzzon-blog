package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/storage"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

const maxUploadSize = 10 << 20 // 10MB

type MediaService struct {
	queries *sqlc.Queries
	storage *storage.S3Storage
}

func NewMediaService(q *sqlc.Queries, s *storage.S3Storage) *MediaService {
	return &MediaService{queries: q, storage: s}
}

func IsAllowedMimeType(mimeType string) bool {
	return allowedMimeTypes[mimeType]
}

func MaxUploadSize() int64 {
	return maxUploadSize
}

func (s *MediaService) Upload(ctx context.Context, body io.Reader, originalName string, mimeType string, sizeBytes int32, width, height int32) (sqlc.Media, error) {
	ext := filepath.Ext(originalName)
	if ext == "" {
		switch mimeType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		}
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), strings.ToLower(ext))
	key := fmt.Sprintf("media/%s", filename)

	url, err := s.storage.Upload(ctx, key, body, mimeType)
	if err != nil {
		return sqlc.Media{}, err
	}

	media, err := s.queries.CreateMedia(ctx, sqlc.CreateMediaParams{
		Filename:     filename,
		OriginalName: originalName,
		URL:          url,
		MimeType:     mimeType,
		SizeBytes:    sizeBytes,
		AltText:      pgtype.Text{},
		Width:        pgtype.Int4{Int32: width, Valid: width > 0},
		Height:       pgtype.Int4{Int32: height, Valid: height > 0},
	})
	if err != nil {
		return sqlc.Media{}, err
	}

	return media, nil
}

func (s *MediaService) Delete(ctx context.Context, id uuid.UUID) error {
	media, err := s.queries.GetMediaByID(ctx, id)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("media/%s", media.Filename)
	if err := s.storage.Delete(ctx, key); err != nil {
		return err
	}

	return s.queries.DeleteMedia(ctx, id)
}
