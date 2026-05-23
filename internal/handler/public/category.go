package public

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/texef-tech/winzzon-blog/internal/cache"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
)

type CategoryHandler struct {
	queries *sqlc.Queries
	cache   *cache.RedisCache
	posts   *PostHandler
}

func NewCategoryHandler(q *sqlc.Queries, c *cache.RedisCache, ph *PostHandler) *CategoryHandler {
	return &CategoryHandler{queries: q, cache: c, posts: ph}
}

type publicCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	if cached, err := h.cache.Get(r.Context(), "categories:list"); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	cats, err := h.queries.ListCategories(r.Context())
	if err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to list categories")
		return
	}

	var resp []publicCategoryResponse
	for _, c := range cats {
		item := publicCategoryResponse{
			ID:   c.ID.String(),
			Name: c.Name,
			Slug: c.Slug,
		}
		if c.Description.Valid {
			item.Description = c.Description.String
		}
		resp = append(resp, item)
	}
	if resp == nil {
		resp = []publicCategoryResponse{}
	}

	if data, err := json.Marshal(resp); err == nil {
		h.cache.Set(r.Context(), "categories:list", string(data), 24*time.Hour)
	}

	handler.JSON(w, http.StatusOK, resp)
}

func (h *CategoryHandler) Posts(w http.ResponseWriter, r *http.Request) {
	categorySlug := chi.URLParam(r, "slug")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	posts, err := h.queries.ListPublishedPostsByCategory(r.Context(), categorySlug, int32(limit), int32(offset))
	if err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to list posts")
		return
	}

	total, err := h.queries.CountPublishedPostsByCategory(r.Context(), categorySlug)
	if err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to count posts by category")
		return
	}

	var items []postListItem
	for _, p := range posts {
		item, err := h.posts.toListItem(r.Context(), p)
		if err != nil {
			handler.ServerError(w, "INTERNAL_ERROR", "Failed to retrieve post details")
			return
		}
		items = append(items, item)
	}
	if items == nil {
		items = []postListItem{}
	}

	handler.JSON(w, http.StatusOK, map[string]interface{}{
		"data":  items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
