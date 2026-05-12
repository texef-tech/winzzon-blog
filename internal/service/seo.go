package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
)

type SEOService struct {
	queries  *sqlc.Queries
	siteURL  string
	siteName string
}

func NewSEOService(q *sqlc.Queries, siteURL, siteName string) *SEOService {
	return &SEOService{queries: q, siteURL: siteURL, siteName: siteName}
}

type URLSet struct {
	XMLName xml.Name  `xml:"urlset"`
	XMLNS   string    `xml:"xmlns,attr"`
	URLs    []SiteURL `xml:"url"`
}

type SiteURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

func (s *SEOService) GenerateSitemap(ctx context.Context) ([]byte, error) {
	posts, err := s.queries.ListAllPublishedPostsForSitemap(ctx)
	if err != nil {
		return nil, err
	}

	urlSet := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	for _, p := range posts {
		urlSet.URLs = append(urlSet.URLs, SiteURL{
			Loc:        fmt.Sprintf("%s/blog/%s", s.siteURL, p.Slug),
			LastMod:    p.UpdatedAt.Format(time.RFC3339),
			ChangeFreq: "weekly",
			Priority:   "0.8",
		})
	}

	data, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), data...), nil
}

type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func (s *SEOService) GenerateRSS(ctx context.Context) ([]byte, error) {
	posts, err := s.queries.ListRecentPublishedPosts(ctx, 20)
	if err != nil {
		return nil, err
	}

	feed := RSSFeed{
		Version: "2.0",
		Channel: RSSChannel{
			Title:       s.siteName + " Blog",
			Link:        s.siteURL,
			Description: fmt.Sprintf("Latest posts from %s", s.siteName),
		},
	}

	for _, p := range posts {
		pubDate := ""
		if p.PublishedAt.Valid {
			pubDate = p.PublishedAt.Time.Format(time.RFC1123Z)
		}

		excerpt := ""
		if p.Excerpt.Valid {
			excerpt = p.Excerpt.String
		}

		feed.Channel.Items = append(feed.Channel.Items, RSSItem{
			Title:       p.Title,
			Link:        fmt.Sprintf("%s/blog/%s", s.siteURL, p.Slug),
			Description: excerpt,
			PubDate:     pubDate,
			GUID:        p.ID.String(),
		})
	}

	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), data...), nil
}

func (s *SEOService) GenerateRobotsTxt() string {
	return fmt.Sprintf("User-agent: *\nDisallow: /api/\nDisallow: /api/v1/admin/\nSitemap: %s/sitemap.xml\n", s.siteURL)
}

type SchemaOrg map[string]interface{}

func (s *SEOService) BuildArticleSchema(post sqlc.Post, seo *sqlc.SeoMeta) SchemaOrg {
	schema := SchemaOrg{
		"@context":      "https://schema.org",
		"@type":         "Article",
		"headline":      post.Title,
		"url":           fmt.Sprintf("%s/blog/%s", s.siteURL, post.Slug),
		"publisher": SchemaOrg{
			"@type": "Organization",
			"name":  s.siteName,
			"url":   s.siteURL,
		},
	}

	if post.PublishedAt.Valid {
		schema["datePublished"] = post.PublishedAt.Time.Format(time.RFC3339)
	}
	if post.Excerpt.Valid {
		schema["description"] = post.Excerpt.String
	}

	if seo != nil {
		if seo.SchemaType.Valid && seo.SchemaType.String != "" {
			schema["@type"] = seo.SchemaType.String
		}
		if seo.OgImage.Valid && seo.OgImage.String != "" {
			schema["image"] = seo.OgImage.String
		}
	}

	return schema
}
