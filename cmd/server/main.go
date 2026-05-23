package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/texef-tech/winzzon-blog/internal/cache"
	"github.com/texef-tech/winzzon-blog/internal/config"
	"github.com/texef-tech/winzzon-blog/internal/db"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	adminHandler "github.com/texef-tech/winzzon-blog/internal/handler/admin"
	publicHandler "github.com/texef-tech/winzzon-blog/internal/handler/public"
	"github.com/texef-tech/winzzon-blog/internal/middleware"
	"github.com/texef-tech/winzzon-blog/internal/service"
	"github.com/texef-tech/winzzon-blog/internal/storage"
)

func main() {
	cfg := config.Load()

	if cfg.IsDevelopment() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()
	log.Info().Msg("connected to PostgreSQL")

	redisCache, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer redisCache.Close()
	log.Info().Msg("connected to Redis")

	queries := sqlc.New(pool)

	s3Store := storage.NewS3Storage(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3CDNURL)

	var searchSvc *service.SearchService
	if cfg.MeiliURL != "" && cfg.MeiliAPIKey != "" {
		searchSvc = service.NewSearchService(cfg.MeiliURL, cfg.MeiliAPIKey)
		log.Info().Msg("connected to Meilisearch")
	}

	postSvc := service.NewPostService(queries, redisCache, searchSvc)
	mediaSvc := service.NewMediaService(queries, s3Store)
	seoSvc := service.NewSEOService(queries, cfg.SiteURL, cfg.SiteName)

	rateLimiter, err := middleware.NewRateLimiter(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create rate limiter")
	}

	authH := adminHandler.NewAuthHandler(queries, cfg.JWTSecret, cfg.JWTExpiryHours)
	adminPostH := adminHandler.NewPostHandler(queries, postSvc)
	adminSeoH := adminHandler.NewSEOHandler(queries)
	adminMediaH := adminHandler.NewMediaHandler(queries, mediaSvc)
	adminCatH := adminHandler.NewCategoryHandler(queries, redisCache)
	adminTagH := adminHandler.NewTagHandler(queries, redisCache)

	pubPostH := publicHandler.NewPostHandler(queries, redisCache, seoSvc)
	pubCatH := publicHandler.NewCategoryHandler(queries, redisCache, pubPostH)
	pubTagH := publicHandler.NewTagHandler(queries, redisCache, pubPostH)
	pubSeoH := publicHandler.NewSEOHandler(seoSvc, redisCache)

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/sitemap.xml", pubSeoH.Sitemap)
	r.Get("/rss.xml", pubSeoH.RSS)
	r.Get("/robots.txt", pubSeoH.RobotsTxt)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter.Limit(60, time.Minute))

			r.Get("/posts", pubPostH.List)
			r.Get("/posts/{slug}", pubPostH.GetBySlug)
			r.Get("/categories", pubCatH.List)
			r.Get("/categories/{slug}/posts", pubCatH.Posts)
			r.Get("/tags", pubTagH.List)
			r.Get("/tags/{slug}/posts", pubTagH.Posts)

			if searchSvc != nil {
				searchH := publicHandler.NewSearchHandler(searchSvc)
				r.Get("/search", searchH.Search)
			}
		})

		r.Route("/admin", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(rateLimiter.Limit(5, time.Minute))
				r.Post("/auth/login", authH.Login)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.JWTAuth(cfg.JWTSecret))
				r.Use(rateLimiter.Limit(20, time.Minute))

				r.Post("/auth/refresh", authH.Refresh)

				r.Get("/posts", adminPostH.List)
				r.Get("/posts/{id}", adminPostH.GetByID)
				r.Post("/posts", adminPostH.Create)
				r.Put("/posts/{id}", adminPostH.Update)
				r.Patch("/posts/{id}/publish", adminPostH.Publish)
				r.Patch("/posts/{id}/unpublish", adminPostH.Unpublish)
				r.Delete("/posts/{id}", adminPostH.Delete)

				r.Get("/posts/{id}/seo", adminSeoH.Get)
				r.Put("/posts/{id}/seo", adminSeoH.Upsert)

				r.Post("/media/upload", adminMediaH.Upload)
				r.Get("/media", adminMediaH.List)
				r.Delete("/media/{id}", adminMediaH.Delete)

				r.Post("/categories", adminCatH.Create)
				r.Put("/categories/{id}", adminCatH.Update)
				r.Delete("/categories/{id}", adminCatH.Delete)

				r.Post("/tags", adminTagH.Create)
				r.Put("/tags/{id}", adminTagH.Update)
				r.Delete("/tags/{id}", adminTagH.Delete)
			})
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().
			Str("port", cfg.Port).
			Str("env", cfg.Env).
			Str("allowed_origins", cfg.AllowedOrigins).
			Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced shutdown")
	}

	fmt.Println("server stopped gracefully")
}
