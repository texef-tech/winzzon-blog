package sanitizer

import (
	"github.com/microcosm-cc/bluemonday"
)

var policy *bluemonday.Policy

func init() {
	policy = bluemonday.NewPolicy()

	policy.AllowElements(
		"p", "h2", "h3", "h4", "h5", "h6",
		"strong", "em", "br", "hr",
		"ul", "ol", "li",
		"blockquote", "code", "pre",
		"table", "thead", "tbody", "tr", "td", "th",
	)

	policy.AllowAttrs("href").OnElements("a")
	policy.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	policy.RequireParseableURLs(true)
}

func Sanitize(html string) string {
	return policy.Sanitize(html)
}

func StripAll(html string) string {
	return bluemonday.StrictPolicy().Sanitize(html)
}
