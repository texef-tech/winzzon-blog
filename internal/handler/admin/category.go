package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/texef-tech/winzzon-blog/internal/cache"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
	"github.com/texef-tech/winzzon-blog/pkg/slug"
)

type CategoryHandler struct {
	queries *sqlc.Queries
	cache   *cache.RedisCache
}

func NewCategoryHandler(q *sqlc.Queries, c *cache.RedisCache) *CategoryHandler {
	return &CategoryHandler{queries: q, cache: c}
}

type categoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type categoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func toCategoryResponse(c sqlc.Category) categoryResponse {
	resp := categoryResponse{
		ID:        c.ID.String(),
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.Description.Valid {
		resp.Description = c.Description.String
	}
	return resp
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		handler.ErrorJSON(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Name is required")
		return
	}

	cat, err := h.queries.CreateCategory(r.Context(), sqlc.CreateCategoryParams{
		Name:        req.Name,
		Slug:        slug.Generate(req.Name),
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		handler.ErrorJSON(w, http.StatusConflict, "CATEGORY_EXISTS", "A category with this name already exists")
		return
	}

	h.cache.Delete(r.Context(), "categories:list")
	handler.JSON(w, http.StatusCreated, toCategoryResponse(cat))
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	cat, err := h.queries.UpdateCategory(r.Context(), sqlc.UpdateCategoryParams{
		ID:          id,
		Name:        req.Name,
		Slug:        slug.Generate(req.Name),
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category")
		return
	}

	h.cache.Delete(r.Context(), "categories:list")
	handler.JSON(w, http.StatusOK, toCategoryResponse(cat))
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	if err := h.queries.DeleteCategory(r.Context(), id); err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category")
		return
	}

	h.cache.Delete(r.Context(), "categories:list")
	w.WriteHeader(http.StatusNoContent)
}
