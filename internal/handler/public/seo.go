package public

import (
	"net/http"
	"time"

	"github.com/texef-tech/winzzon-blog/internal/cache"
	"github.com/texef-tech/winzzon-blog/internal/service"
)

type SEOHandler struct {
	seoService *service.SEOService
	cache      *cache.RedisCache
}

func NewSEOHandler(seo *service.SEOService, c *cache.RedisCache) *SEOHandler {
	return &SEOHandler{seoService: seo, cache: c}
}

func (h *SEOHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	if cached, err := h.cache.Get(r.Context(), "sitemap:xml"); err == nil {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(cached))
		return
	}

	data, err := h.seoService.GenerateSitemap(r.Context())
	if err != nil {
		http.Error(w, "Failed to generate sitemap", http.StatusInternalServerError)
		return
	}

	h.cache.Set(r.Context(), "sitemap:xml", string(data), 60*time.Minute)

	w.Header().Set("Content-Type", "application/xml")
	w.Write(data)
}

func (h *SEOHandler) RSS(w http.ResponseWriter, r *http.Request) {
	if cached, err := h.cache.Get(r.Context(), "rss:xml"); err == nil {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(cached))
		return
	}

	data, err := h.seoService.GenerateRSS(r.Context())
	if err != nil {
		http.Error(w, "Failed to generate RSS feed", http.StatusInternalServerError)
		return
	}

	h.cache.Set(r.Context(), "rss:xml", string(data), 30*time.Minute)

	w.Header().Set("Content-Type", "application/rss+xml")
	w.Write(data)
}

func (h *SEOHandler) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(h.seoService.GenerateRobotsTxt()))
}
