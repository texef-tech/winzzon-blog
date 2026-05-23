package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
)

type SEOHandler struct {
	queries *sqlc.Queries
}

func NewSEOHandler(q *sqlc.Queries) *SEOHandler {
	return &SEOHandler{queries: q}
}

type seoRequest struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	OgTitle         string `json:"og_title"`
	OgDescription   string `json:"og_description"`
	OgImage         string `json:"og_image"`
	CanonicalURL    string `json:"canonical_url"`
	FocusKeyword    string `json:"focus_keyword"`
	SchemaType      string `json:"schema_type"`
	NoIndex         bool   `json:"no_index"`
}

type seoResponse struct {
	ID              string `json:"id"`
	PostID          string `json:"post_id"`
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	OgTitle         string `json:"og_title,omitempty"`
	OgDescription   string `json:"og_description,omitempty"`
	OgImage         string `json:"og_image,omitempty"`
	CanonicalURL    string `json:"canonical_url,omitempty"`
	FocusKeyword    string `json:"focus_keyword,omitempty"`
	SchemaType      string `json:"schema_type,omitempty"`
	NoIndex         bool   `json:"no_index"`
}

func toSEOResponse(s sqlc.SeoMeta) seoResponse {
	resp := seoResponse{
		ID:     s.ID.String(),
		PostID: s.PostID.String(),
	}
	if s.MetaTitle.Valid {
		resp.MetaTitle = s.MetaTitle.String
	}
	if s.MetaDescription.Valid {
		resp.MetaDescription = s.MetaDescription.String
	}
	if s.OgTitle.Valid {
		resp.OgTitle = s.OgTitle.String
	}
	if s.OgDescription.Valid {
		resp.OgDescription = s.OgDescription.String
	}
	if s.OgImage.Valid {
		resp.OgImage = s.OgImage.String
	}
	if s.CanonicalURL.Valid {
		resp.CanonicalURL = s.CanonicalURL.String
	}
	if s.FocusKeyword.Valid {
		resp.FocusKeyword = s.FocusKeyword.String
	}
	if s.SchemaType.Valid {
		resp.SchemaType = s.SchemaType.String
	}
	if s.NoIndex.Valid {
		resp.NoIndex = s.NoIndex.Bool
	}
	return resp
}

func (h *SEOHandler) Get(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.BadRequest(w, "INVALID_ID", "Invalid post ID")
		return
	}

	seo, err := h.queries.GetSEOByPostID(r.Context(), postID)
	if err != nil {
		handler.NotFound(w, "SEO_NOT_FOUND", "SEO meta not found for this post")
		return
	}

	handler.JSON(w, http.StatusOK, toSEOResponse(seo))
}

func (h *SEOHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.BadRequest(w, "INVALID_ID", "Invalid post ID")
		return
	}

	var req seoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.BadRequest(w, "INVALID_BODY", "Invalid request body")
		return
	}

	if len(req.MetaTitle) > 60 {
		handler.ValidationErrorJSON(w, map[string]string{"meta_title": "Meta title must be at most 60 characters"})
		return
	}
	if len(req.MetaDescription) > 160 {
		handler.ValidationErrorJSON(w, map[string]string{"meta_description": "Meta description must be at most 160 characters"})
		return
	}

	seo, err := h.queries.UpsertSEO(r.Context(), sqlc.UpsertSEOParams{
		PostID:          postID,
		MetaTitle:       pgtype.Text{String: req.MetaTitle, Valid: req.MetaTitle != ""},
		MetaDescription: pgtype.Text{String: req.MetaDescription, Valid: req.MetaDescription != ""},
		OgTitle:         pgtype.Text{String: req.OgTitle, Valid: req.OgTitle != ""},
		OgDescription:   pgtype.Text{String: req.OgDescription, Valid: req.OgDescription != ""},
		OgImage:         pgtype.Text{String: req.OgImage, Valid: req.OgImage != ""},
		CanonicalURL:    pgtype.Text{String: req.CanonicalURL, Valid: req.CanonicalURL != ""},
		FocusKeyword:    pgtype.Text{String: req.FocusKeyword, Valid: req.FocusKeyword != ""},
		SchemaType:      pgtype.Text{String: req.SchemaType, Valid: req.SchemaType != ""},
		NoIndex:         pgtype.Bool{Bool: req.NoIndex, Valid: true},
	})
	if err != nil {
		handler.ServerError(w, "INTERNAL_ERROR", "Failed to save SEO meta")
		return
	}

	handler.JSON(w, http.StatusOK, toSEOResponse(seo))
}
