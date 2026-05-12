package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/texef-tech/winzzon-blog/internal/cache"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/pkg/sanitizer"
	"github.com/texef-tech/winzzon-blog/pkg/slug"
)

type PostService struct {
	queries *sqlc.Queries
	cache   *cache.RedisCache
	search  *SearchService
}

func NewPostService(q *sqlc.Queries, c *cache.RedisCache, s *SearchService) *PostService {
	return &PostService{queries: q, cache: c, search: s}
}

type CreatePostInput struct {
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Excerpt     string    `json:"excerpt"`
	CategoryIDs []string  `json:"category_ids"`
	TagIDs      []string  `json:"tag_ids"`
}

type UpdatePostInput struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Excerpt     string   `json:"excerpt"`
	CategoryIDs []string `json:"category_ids"`
	TagIDs      []string `json:"tag_ids"`
}

func (s *PostService) Create(ctx context.Context, input CreatePostInput) (sqlc.Post, error) {
	postSlug := slug.Generate(input.Title)
	exists, err := s.queries.SlugExists(ctx, postSlug)
	if err != nil {
		return sqlc.Post{}, err
	}
	if exists {
		for i := 2; ; i++ {
			candidate := fmt.Sprintf("%s-%d", postSlug, i)
			exists, err = s.queries.SlugExists(ctx, candidate)
			if err != nil {
				return sqlc.Post{}, err
			}
			if !exists {
				postSlug = candidate
				break
			}
		}
	}

	sanitizedBody := sanitizer.Sanitize(input.Body)
	plainText := sanitizer.StripAll(input.Body)
	words := len(strings.Fields(plainText))
	readingTime := words / 200

	post, err := s.queries.CreatePost(ctx, sqlc.CreatePostParams{
		Title:       input.Title,
		Slug:        postSlug,
		Body:        sanitizedBody,
		BodyPlain:   pgtype.Text{String: plainText, Valid: true},
		Excerpt:     pgtype.Text{String: input.Excerpt, Valid: input.Excerpt != ""},
		ReadingTime: pgtype.Int4{Int32: int32(readingTime), Valid: true},
		Status:      "draft",
	})
	if err != nil {
		return sqlc.Post{}, err
	}

	s.syncPostRelations(ctx, post.ID, input.CategoryIDs, input.TagIDs)

	return post, nil
}

func (s *PostService) Update(ctx context.Context, id uuid.UUID, input UpdatePostInput) (sqlc.Post, error) {
	sanitizedBody := sanitizer.Sanitize(input.Body)
	plainText := sanitizer.StripAll(input.Body)
	words := len(strings.Fields(plainText))
	readingTime := words / 200

	post, err := s.queries.UpdatePost(ctx, sqlc.UpdatePostParams{
		ID:          id,
		Title:       input.Title,
		Body:        sanitizedBody,
		BodyPlain:   pgtype.Text{String: plainText, Valid: true},
		Excerpt:     pgtype.Text{String: input.Excerpt, Valid: input.Excerpt != ""},
		ReadingTime: pgtype.Int4{Int32: int32(readingTime), Valid: true},
	})
	if err != nil {
		return sqlc.Post{}, err
	}

	s.syncPostRelations(ctx, post.ID, input.CategoryIDs, input.TagIDs)
	s.invalidatePostCache(ctx, post.Slug)

	if post.Status == "published" && s.search != nil {
		_ = s.search.IndexPost(ctx, post)
	}

	return post, nil
}

func (s *PostService) Publish(ctx context.Context, id uuid.UUID) (sqlc.Post, error) {
	post, err := s.queries.PublishPost(ctx, id)
	if err != nil {
		return sqlc.Post{}, err
	}

	s.invalidatePostCache(ctx, post.Slug)
	s.cache.Delete(ctx, "sitemap:xml", "rss:xml")
	s.cache.DeleteByPattern(ctx, "posts:list:*")

	if s.search != nil {
		_ = s.search.IndexPost(ctx, post)
	}

	return post, nil
}

func (s *PostService) Unpublish(ctx context.Context, id uuid.UUID) (sqlc.Post, error) {
	post, err := s.queries.UnpublishPost(ctx, id)
	if err != nil {
		return sqlc.Post{}, err
	}

	s.invalidatePostCache(ctx, post.Slug)
	s.cache.Delete(ctx, "sitemap:xml", "rss:xml")
	s.cache.DeleteByPattern(ctx, "posts:list:*")

	if s.search != nil {
		_ = s.search.RemovePost(ctx, post.ID.String())
	}

	return post, nil
}

func (s *PostService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	post, err := s.queries.GetPostByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.queries.SoftDeletePost(ctx, id); err != nil {
		return err
	}

	s.invalidatePostCache(ctx, post.Slug)
	s.cache.DeleteByPattern(ctx, "posts:list:*")

	if post.Status == "published" {
		s.cache.Delete(ctx, "sitemap:xml", "rss:xml")
		if s.search != nil {
			_ = s.search.RemovePost(ctx, post.ID.String())
		}
	}

	return nil
}

func (s *PostService) syncPostRelations(ctx context.Context, postID uuid.UUID, categoryIDs, tagIDs []string) {
	_ = s.queries.RemovePostCategories(ctx, postID)
	for _, cid := range categoryIDs {
		id, err := uuid.Parse(cid)
		if err != nil {
			continue
		}
		_ = s.queries.AddPostCategory(ctx, postID, id)
	}

	_ = s.queries.RemovePostTags(ctx, postID)
	for _, tid := range tagIDs {
		id, err := uuid.Parse(tid)
		if err != nil {
			continue
		}
		_ = s.queries.AddPostTag(ctx, postID, id)
	}
}

func (s *PostService) invalidatePostCache(ctx context.Context, postSlug string) {
	s.cache.Delete(ctx, fmt.Sprintf("post:slug:%s", postSlug))
	s.cache.DeleteByPattern(ctx, "posts:list:*")
}
