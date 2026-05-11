package main

import (
	"context"
	"fmt"
	"log"

	"github.com/texef-tech/winzzon-blog/internal/config"
	"github.com/texef-tech/winzzon-blog/internal/db"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/service"
)

func main() {
	cfg := config.Load()

	if cfg.MeiliURL == "" || cfg.MeiliAPIKey == "" {
		log.Fatal("MEILI_URL and MEILI_API_KEY must be set")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	searchSvc := service.NewSearchService(cfg.MeiliURL, cfg.MeiliAPIKey)

	posts, err := queries.ListRecentPublishedPosts(ctx, 10000)
	if err != nil {
		log.Fatalf("failed to list posts: %v", err)
	}

	fmt.Printf("reindexing %d published posts...\n", len(posts))

	for i, post := range posts {
		if err := searchSvc.IndexPost(ctx, post); err != nil {
			fmt.Printf("  [%d/%d] FAILED %s: %v\n", i+1, len(posts), post.Slug, err)
		} else {
			fmt.Printf("  [%d/%d] indexed %s\n", i+1, len(posts), post.Slug)
		}
	}

	fmt.Println("reindex complete")
}
