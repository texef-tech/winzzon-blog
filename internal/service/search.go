package service

import (
	"context"
	"encoding/json"

	"github.com/meilisearch/meilisearch-go"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
)

type SearchService struct {
	client meilisearch.ServiceManager
	index  string
}

func NewSearchService(meiliURL, apiKey string) *SearchService {
	client := meilisearch.New(meiliURL, meilisearch.WithAPIKey(apiKey))

	idx := "posts"
	client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        idx,
		PrimaryKey: "id",
	})

	client.Index(idx).UpdateSearchableAttributes(&[]string{
		"title", "body_plain", "excerpt", "tags", "categories",
	})

	return &SearchService{client: client, index: idx}
}

type SearchDocument struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	BodyPlain     string   `json:"body_plain"`
	Excerpt       string   `json:"excerpt"`
	ReadingTime   int32    `json:"reading_time"`
	PublishedAt   string   `json:"published_at"`
	CategorySlugs []string `json:"category_slugs"`
	TagSlugs      []string `json:"tag_slugs"`
	OgImage       string   `json:"og_image"`
}

func (s *SearchService) IndexPost(ctx context.Context, post sqlc.Post) error {
	doc := SearchDocument{
		ID:    post.ID.String(),
		Title: post.Title,
		Slug:  post.Slug,
	}
	if post.BodyPlain.Valid {
		doc.BodyPlain = post.BodyPlain.String
	}
	if post.Excerpt.Valid {
		doc.Excerpt = post.Excerpt.String
	}
	if post.ReadingTime.Valid {
		doc.ReadingTime = post.ReadingTime.Int32
	}
	if post.PublishedAt.Valid {
		doc.PublishedAt = post.PublishedAt.Time.String()
	}

	pk := "id"
	_, err := s.client.Index(s.index).AddDocuments([]SearchDocument{doc}, &meilisearch.DocumentOptions{
		PrimaryKey: &pk,
	})
	return err
}

func (s *SearchService) RemovePost(ctx context.Context, id string) error {
	_, err := s.client.Index(s.index).DeleteDocument(id, nil)
	return err
}

type SearchResult struct {
	Hits             []map[string]interface{} `json:"hits"`
	EstimatedTotal   int64                    `json:"estimated_total"`
	ProcessingTimeMs int64                    `json:"processing_time_ms"`
}

func (s *SearchService) Search(ctx context.Context, query string, limit, offset int64) (*SearchResult, error) {
	resp, err := s.client.Index(s.index).Search(query, &meilisearch.SearchRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	hits := make([]map[string]interface{}, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		m := make(map[string]interface{})
		for k, v := range hit {
			var val interface{}
			if err := json.Unmarshal(v, &val); err == nil {
				m[k] = val
			}
		}
		hits = append(hits, m)
	}

	return &SearchResult{
		Hits:             hits,
		EstimatedTotal:   resp.EstimatedTotalHits,
		ProcessingTimeMs: resp.ProcessingTimeMs,
	}, nil
}
