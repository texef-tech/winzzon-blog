package public

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/texef-tech/winzzon-blog/internal/handler"
	"github.com/texef-tech/winzzon-blog/internal/storage"
)

type MediaHandler struct {
	storage *storage.S3Storage
}

func NewMediaHandler(s *storage.S3Storage) *MediaHandler {
	return &MediaHandler{storage: s}
}

func (h *MediaHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		handler.BadRequest(w, "MISSING_FILENAME", "Filename is required")
		return
	}

	// The key in the bucket is "media/<filename>"
	key := "media/" + filename

	body, contentType, contentLength, err := h.storage.Get(r.Context(), key)
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			handler.NotFound(w, "MEDIA_NOT_FOUND", "Media file not found")
			return
		}
		if strings.Contains(err.Error(), "NoSuchKey") {
			handler.NotFound(w, "MEDIA_NOT_FOUND", "Media file not found")
			return
		}
		log.Error().Err(err).Str("key", key).Msg("failed to fetch media from S3")
		handler.ServerError(w, "FETCH_FAILED", "Failed to retrieve media file")
		return
	}
	defer body.Close()

	// Set standard headers for file download/view
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	if contentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}

	// Enable browser caching for 1 year since blog media is static
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	// Stream the file body to the response writer
	if _, err := io.Copy(w, body); err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to stream media body")
	}
}
