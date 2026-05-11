package public

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/texef-tech/winzzon-blog/internal/cache"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
	"github.com/texef-tech/winzzon-blog/internal/service"
)

type PostHandler struct {
	queries    *sqlc.Queries
	cache      *cache.RedisCache
	seoService *service.SEOService
}

func NewPostHandler(q *sqlc.Queries, c *cache.RedisCache, seo *service.SEOService) *PostHandler {
	return &PostHandler{queries: q, cache: c, seoService: seo}
}

type publicPostResponse struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	Body        string            `json:"body"`
	Excerpt     string            `json:"excerpt,omitempty"`
	ReadingTime int32             `json:"reading_time"`
	PublishedAt string            `json:"published_at,omitempty"`
	Categories  []briefItem       `json:"categories"`
	Tags        []briefItem       `json:"tags"`
	SEO         *seoBlock         `json:"seo,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

type briefItem struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type seoBlock struct {
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	OgImage         string `json:"og_image,omitempty"`
	CanonicalURL    string `json:"canonical_url,omitempty"`
	SchemaType      string `json:"schema_type,omitempty"`
	NoIndex         bool   `json:"no_index"`
}

type postListItem struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Slug        string      `json:"slug"`
	Excerpt     string      `json:"excerpt,omitempty"`
	ReadingTime int32       `json:"reading_time"`
	PublishedAt string      `json:"published_at,omitempty"`
	Categories  []briefItem `json:"categories"`
	Tags        []briefItem `json:"tags"`
}

func (h *PostHandler) toListItem(ctx context.Context, p sqlc.Post) postListItem {
	item := postListItem{
		ID:          p.ID.String(),
		Title:       p.Title,
		Slug:        p.Slug,
		ReadingTime: p.ReadingTime.Int32,
	}
	if p.Excerpt.Valid {
		item.Excerpt = p.Excerpt.String
	}
	if p.PublishedAt.Valid {
		item.PublishedAt = p.PublishedAt.Time.Format(time.RFC3339)
	}

	cats, _ := h.queries.ListCategoriesForPost(ctx, p.ID)
	for _, c := range cats {
		item.Categories = append(item.Categories, briefItem{Name: c.Name, Slug: c.Slug})
	}
	if item.Categories == nil {
		item.Categories = []briefItem{}
	}

	tags, _ := h.queries.ListTagsForPost(ctx, p.ID)
	for _, t := range tags {
		item.Tags = append(item.Tags, briefItem{Name: t.Name, Slug: t.Slug})
	}
	if item.Tags == nil {
		item.Tags = []briefItem{}
	}

	return item
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	cacheKey := fmt.Sprintf("posts:list:%x", sha256.Sum256([]byte(fmt.Sprintf("%d:%d", page, limit))))

	if cached, err := h.cache.Get(r.Context(), cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	posts, err := h.queries.ListPublishedPosts(r.Context(), int32(limit), int32(offset))
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list posts")
		return
	}

	total, _ := h.queries.CountPublishedPosts(r.Context())

	var items []postListItem
	for _, p := range posts {
		items = append(items, h.toListItem(r.Context(), p))
	}
	if items == nil {
		items = []postListItem{}
	}

	resp := map[string]interface{}{
		"data":  items,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	if data, err := json.Marshal(resp); err == nil {
		h.cache.Set(r.Context(), cacheKey, string(data), 10*time.Minute)
	}

	handler.JSON(w, http.StatusOK, resp)
}

func (h *PostHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	cacheKey := fmt.Sprintf("post:slug:%s", slug)
	if cached, err := h.cache.Get(r.Context(), cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	post, err := h.queries.GetPostBySlug(r.Context(), slug)
	if err != nil {
		handler.ErrorJSON(w, http.StatusNotFound, "POST_NOT_FOUND", fmt.Sprintf("No published post found with slug: %s", slug))
		return
	}

	resp := publicPostResponse{
		ID:          post.ID.String(),
		Title:       post.Title,
		Slug:        post.Slug,
		Body:        post.Body,
		ReadingTime: post.ReadingTime.Int32,
	}
	if post.Excerpt.Valid {
		resp.Excerpt = post.Excerpt.String
	}
	if post.PublishedAt.Valid {
		resp.PublishedAt = post.PublishedAt.Time.Format(time.RFC3339)
	}

	cats, _ := h.queries.ListCategoriesForPost(r.Context(), post.ID)
	resp.Categories = []briefItem{}
	for _, c := range cats {
		resp.Categories = append(resp.Categories, briefItem{Name: c.Name, Slug: c.Slug})
	}

	tags, _ := h.queries.ListTagsForPost(r.Context(), post.ID)
	resp.Tags = []briefItem{}
	for _, t := range tags {
		resp.Tags = append(resp.Tags, briefItem{Name: t.Name, Slug: t.Slug})
	}

	seo, err := h.queries.GetSEOByPostID(r.Context(), post.ID)
	if err == nil {
		resp.SEO = &seoBlock{
			NoIndex: seo.NoIndex.Valid && seo.NoIndex.Bool,
		}
		if seo.MetaTitle.Valid {
			resp.SEO.MetaTitle = seo.MetaTitle.String
		}
		if seo.MetaDescription.Valid {
			resp.SEO.MetaDescription = seo.MetaDescription.String
		}
		if seo.OgImage.Valid {
			resp.SEO.OgImage = seo.OgImage.String
		}
		if seo.CanonicalURL.Valid {
			resp.SEO.CanonicalURL = seo.CanonicalURL.String
		}
		if seo.SchemaType.Valid {
			resp.SEO.SchemaType = seo.SchemaType.String
		}
		resp.Schema = h.seoService.BuildArticleSchema(post, &seo)
	} else {
		resp.Schema = h.seoService.BuildArticleSchema(post, nil)
	}

	if data, err := json.Marshal(resp); err == nil {
		h.cache.Set(r.Context(), cacheKey, string(data), 60*time.Minute)
	}

	handler.JSON(w, http.StatusOK, resp)
}
