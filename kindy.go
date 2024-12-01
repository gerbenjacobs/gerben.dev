package gerbendev

import (
	"html/template"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

// Kindy is a datastructure for content that adheres to Microformats 2
type Kindy struct {
	Type        string             `json:"type"`
	MFType      string             `json:"mfType"` // Microformat type; often h-entry
	Title       string             `json:"title,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	PublishedAt time.Time          `json:"publishedAt"`
	Content     template.HTML      `json:"content,omitempty"`
	Slug        string             `json:"slug,omitempty"`
	Permalink   string             `json:"permalink,omitempty"`
	Author      *KindyAuthor       `json:"author,omitempty"`
	Syndication []KindySyndication `json:"syndication,omitempty"`
	LikeOf      string             `json:"likeOf,omitempty"`
	Geo         *KindyGeo          `json:"geo,omitempty"`
}

type KindyAuthor struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Photo string `json:"photo,omitempty"`
}

type KindySyndication struct {
	Type string `json:"type,omitempty"` // Free form field to be set by the user, can be used to display different type of icons
	URL  string `json:"url,omitempty"`
}

type KindyGeo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ContentStripped strips all HTML with a strict policy
// but still returns a template.HTML so that properly escaped HTML entities still work
// It has an 'optional' args list, but we really only except 1 int which limits the length
func (k Kindy) ContentStripped(args ...int) template.HTML {
	content := string(k.Content)
	if len(args) > 0 && len(content) > args[0] {
		content = content[:args[0]] + "&hellip;"
	}
	p := bluemonday.StrictPolicy()
	return template.HTML(p.Sanitize(content))
}

func (k Kindy) MustTitle() string {
	if k.Title != "" {
		return k.Title
	}
	if k.Summary != "" {
		return k.Summary
	}
	if k.Type == "note" {
		// for Notes, we might as well use the content
		p := bluemonday.StrictPolicy()
		return p.Sanitize(string(k.Content))
	}
	if k.Permalink != "" {
		return k.Permalink
	}

	return k.Type
}

func (k Kindy) MustDescription() string {
	if k.Summary != "" {
		return k.Summary
	}
	if k.Content != "" {
		p := bluemonday.StrictPolicy()
		return p.Sanitize(string(k.Content))
	}

	return k.Type
}
