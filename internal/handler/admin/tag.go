package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/texef-tech/winzzon-blog/internal/cache"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
	"github.com/texef-tech/winzzon-blog/pkg/slug"
)

type TagHandler struct {
	queries *sqlc.Queries
	cache   *cache.RedisCache
}

func NewTagHandler(q *sqlc.Queries, c *cache.RedisCache) *TagHandler {
	return &TagHandler{queries: q, cache: c}
}

type tagRequest struct {
	Name string `json:"name"`
}

type tagResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

func toTagResponse(t sqlc.Tag) tagResponse {
	return tagResponse{
		ID:        t.ID.String(),
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.BadRequest(w, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		handler.ValidationErrorJSON(w, map[string]string{"name": "Name is required"})
		return
	}

	tag, err := h.queries.CreateTag(r.Context(), sqlc.CreateTagParams{
		Name: req.Name,
		Slug: slug.Generate(req.Name),
	})
	if err != nil {
		handler.Conflict(w, "TAG_EXISTS", "A tag with this name already exists")
		return
	}

	h.cache.Delete(r.Context(), "tags:list")
	handler.JSON(w, http.StatusCreated, toTagResponse(tag))
}

func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.BadRequest(w, "INVALID_ID", "Invalid tag ID")
		return
	}

	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.BadRequest(w, "INVALID_BODY", "Invalid request body")
		return
	}

	tag, err := h.queries.UpdateTag(r.Context(), sqlc.UpdateTagParams{
		ID:   id,
		Name: req.Name,
		Slug: slug.Generate(req.Name),
	})
	if err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to update tag")
		return
	}

	h.cache.Delete(r.Context(), "tags:list")
	handler.JSON(w, http.StatusOK, toTagResponse(tag))
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.BadRequest(w, "INVALID_ID", "Invalid tag ID")
		return
	}

	if err := h.queries.DeleteTag(r.Context(), id); err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to delete tag")
		return
	}

	h.cache.Delete(r.Context(), "tags:list")
	w.WriteHeader(http.StatusNoContent)
}
