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

type TagHandler struct {
	queries *sqlc.Queries
	cache   *cache.RedisCache
	posts   *PostHandler
}

func NewTagHandler(q *sqlc.Queries, c *cache.RedisCache, ph *PostHandler) *TagHandler {
	return &TagHandler{queries: q, cache: c, posts: ph}
}

type publicTagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	if cached, err := h.cache.Get(r.Context(), "tags:list"); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	tags, err := h.queries.ListTags(r.Context())
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tags")
		return
	}

	var resp []publicTagResponse
	for _, t := range tags {
		resp = append(resp, publicTagResponse{
			ID:   t.ID.String(),
			Name: t.Name,
			Slug: t.Slug,
		})
	}
	if resp == nil {
		resp = []publicTagResponse{}
	}

	if data, err := json.Marshal(resp); err == nil {
		h.cache.Set(r.Context(), "tags:list", string(data), 24*time.Hour)
	}

	handler.JSON(w, http.StatusOK, resp)
}

func (h *TagHandler) Posts(w http.ResponseWriter, r *http.Request) {
	tagSlug := chi.URLParam(r, "slug")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	posts, err := h.queries.ListPublishedPostsByTag(r.Context(), tagSlug, int32(limit), int32(offset))
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list posts")
		return
	}

	total, _ := h.queries.CountPublishedPostsByTag(r.Context(), tagSlug)

	var items []postListItem
	for _, p := range posts {
		items = append(items, h.posts.toListItem(r.Context(), p))
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
