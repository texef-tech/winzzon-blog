package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
	"github.com/texef-tech/winzzon-blog/internal/service"
)

type PostHandler struct {
	queries     *sqlc.Queries
	postService *service.PostService
}

func NewPostHandler(q *sqlc.Queries, ps *service.PostService) *PostHandler {
	return &PostHandler{queries: q, postService: ps}
}

type postResponse struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Slug        string      `json:"slug"`
	Body        string      `json:"body"`
	Excerpt     string      `json:"excerpt,omitempty"`
	ReadingTime int32       `json:"reading_time"`
	Status      string      `json:"status"`
	PublishedAt *string     `json:"published_at,omitempty"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	Categories  []catBrief  `json:"categories,omitempty"`
	Tags        []tagBrief  `json:"tags,omitempty"`
}

type catBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type tagBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func toPostResponse(p sqlc.Post) postResponse {
	resp := postResponse{
		ID:          p.ID.String(),
		Title:       p.Title,
		Slug:        p.Slug,
		Body:        p.Body,
		ReadingTime: p.ReadingTime.Int32,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.Excerpt.Valid {
		resp.Excerpt = p.Excerpt.String
	}
	if p.PublishedAt.Valid {
		s := p.PublishedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.PublishedAt = &s
	}
	return resp
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

	posts, err := h.queries.ListAllPosts(r.Context(), int32(limit), int32(offset))
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list posts")
		return
	}

	total, _ := h.queries.CountAllPosts(r.Context())

	var items []postResponse
	for _, p := range posts {
		items = append(items, toPostResponse(p))
	}
	if items == nil {
		items = []postResponse{}
	}

	handler.JSON(w, http.StatusOK, map[string]interface{}{
		"data":  items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_ID", "Invalid post ID")
		return
	}

	post, err := h.queries.GetPostByID(r.Context(), id)
	if err != nil {
		handler.ErrorJSON(w, http.StatusNotFound, "POST_NOT_FOUND", "Post not found")
		return
	}

	resp := toPostResponse(post)
	cats, _ := h.queries.ListCategoriesForPost(r.Context(), post.ID)
	tags, _ := h.queries.ListTagsForPost(r.Context(), post.ID)
	for _, c := range cats {
		resp.Categories = append(resp.Categories, catBrief{ID: c.ID.String(), Name: c.Name, Slug: c.Slug})
	}
	for _, t := range tags {
		resp.Tags = append(resp.Tags, tagBrief{ID: t.ID.String(), Name: t.Name, Slug: t.Slug})
	}

	handler.JSON(w, http.StatusOK, resp)
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreatePostInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if len(input.Title) < 3 || len(input.Title) > 255 {
		handler.ErrorJSON(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Title must be between 3 and 255 characters")
		return
	}
	if len(input.Body) < 10 {
		handler.ErrorJSON(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Body must be at least 10 characters")
		return
	}
	if len(input.Excerpt) > 300 {
		handler.ErrorJSON(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Excerpt must be at most 300 characters")
		return
	}

	post, err := h.postService.Create(r.Context(), input)
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create post")
		return
	}

	handler.JSON(w, http.StatusCreated, toPostResponse(post))
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_ID", "Invalid post ID")
		return
	}

	var input service.UpdatePostInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if len(input.Title) < 3 || len(input.Title) > 255 {
		handler.ErrorJSON(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Title must be between 3 and 255 characters")
		return
	}
	if len(input.Body) < 10 {
		handler.ErrorJSON(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Body must be at least 10 characters")
		return
	}

	post, err := h.postService.Update(r.Context(), id, input)
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update post")
		return
	}

	handler.JSON(w, http.StatusOK, toPostResponse(post))
}

func (h *PostHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_ID", "Invalid post ID")
		return
	}

	post, err := h.postService.Publish(r.Context(), id)
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to publish post")
		return
	}

	handler.JSON(w, http.StatusOK, toPostResponse(post))
}

func (h *PostHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_ID", "Invalid post ID")
		return
	}

	post, err := h.postService.Unpublish(r.Context(), id)
	if err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unpublish post")
		return
	}

	handler.JSON(w, http.StatusOK, toPostResponse(post))
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.ErrorJSON(w, http.StatusBadRequest, "INVALID_ID", "Invalid post ID")
		return
	}

	if err := h.postService.SoftDelete(r.Context(), id); err != nil {
		handler.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
