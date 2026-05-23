package admin

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
	"github.com/texef-tech/winzzon-blog/internal/service"
)

type MediaHandler struct {
	queries      *sqlc.Queries
	mediaService *service.MediaService
}

func NewMediaHandler(q *sqlc.Queries, ms *service.MediaService) *MediaHandler {
	return &MediaHandler{queries: q, mediaService: ms}
}

type mediaResponse struct {
	ID           string `json:"id"`
	Filename     string `json:"filename"`
	OriginalName string `json:"original_name"`
	URL          string `json:"url"`
	MimeType     string `json:"mime_type"`
	SizeBytes    int32  `json:"size_bytes"`
	AltText      string `json:"alt_text,omitempty"`
	Width        int32  `json:"width,omitempty"`
	Height       int32  `json:"height,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func toMediaResponse(m sqlc.Media) mediaResponse {
	resp := mediaResponse{
		ID:           m.ID.String(),
		Filename:     m.Filename,
		OriginalName: m.OriginalName,
		URL:          m.URL,
		MimeType:     m.MimeType,
		SizeBytes:    m.SizeBytes,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if m.AltText.Valid {
		resp.AltText = m.AltText.String
	}
	if m.Width.Valid {
		resp.Width = m.Width.Int32
	}
	if m.Height.Valid {
		resp.Height = m.Height.Int32
	}
	return resp
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxUploadSize())

	if err := r.ParseMultipartForm(service.MaxUploadSize()); err != nil {
		handler.ErrorJSON(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Upload exceeds 10MB limit")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		handler.BadRequest(w, "MISSING_FILE", "File field is required")
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if !service.IsAllowedMimeType(contentType) {
		handler.Unprocessable(w, "INVALID_FILE_TYPE", "Only jpeg, png, webp, and gif images are allowed")
		return
	}

	media, err := h.mediaService.Upload(
		r.Context(), file, fileHeader.Filename, contentType, int32(fileHeader.Size), 0, 0,
	)
	if err != nil {
		handler.ServerError(w, "UPLOAD_FAILED", "Failed to upload file")
		return
	}

	handler.JSON(w, http.StatusCreated, toMediaResponse(media))
}

func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	items, err := h.queries.ListMedia(r.Context(), int32(limit), int32(offset))
	if err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to list media")
		return
	}

	total, err := h.queries.CountMedia(r.Context())
	if err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to count media")
		return
	}

	var resp []mediaResponse
	for _, m := range items {
		resp = append(resp, toMediaResponse(m))
	}
	if resp == nil {
		resp = []mediaResponse{}
	}

	handler.JSON(w, http.StatusOK, map[string]interface{}{
		"data":  resp,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.BadRequest(w, "INVALID_ID", "Invalid media ID")
		return
	}

	if err := h.mediaService.Delete(r.Context(), id); err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to delete media")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
