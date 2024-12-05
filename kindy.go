package gerbendev

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

type KindyType string

const (
	KindyTypeNote   = "note"
	KindyTypePost   = "post"
	KindyTypePhoto  = "photo"
	KindyTypeLike   = "like"
	KindyTypeRepost = "repost"
)

// Kindy is a datastructure for content that adheres to Microformats 2
type Kindy struct {
	Type        KindyType          `json:"type"`
	Title       string             `json:"title,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	PublishedAt time.Time          `json:"publishedAt"`
	Content     template.HTML      `json:"content,omitempty"`
	Slug        string             `json:"slug,omitempty"`
	Permalink   string             `json:"permalink,omitempty"`
	Author      *KindyAuthor       `json:"author,omitempty"`
	Syndication []KindySyndication `json:"syndication,omitempty"`
	LikeOf      string             `json:"likeOf,omitempty"`
	RepostOf    string             `json:"repostOf,omitempty"`
	Geo         *KindyGeo          `json:"geo,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
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

// MFType returns the microformat h-type based on the type of Kindy
func (k Kindy) MFType() string {
	return "h-entry"
}

func (k Kindy) TypeEmoji() string {
	emojis := map[KindyType]string{
		KindyTypePost:   "📝",
		KindyTypeNote:   "🗒",
		KindyTypePhoto:  "📸",
		KindyTypeRepost: "🔁",
		KindyTypeLike:   "⭐",
	}
	if v, ok := emojis[k.Type]; ok {
		return v
	}

	return emojis[KindyTypePost]
}

func (k Kindy) Thumbnail() string {
	if k.Type == KindyTypePhoto {
		filePath := string(k.Content)
		ext := filepath.Ext(filePath)
		return fmt.Sprintf("%s_thumb%s", strings.TrimSuffix(filePath, ext), ext)
	}

	return ""
}

// ContentStripped strips all HTML with a strict policy
// but still returns a template.HTML so that properly escaped HTML entities still work
// It has an 'optional' args list, but we really only except 1 int which limits the length
func (k Kindy) ContentStripped(args ...int) template.HTML {
	content := strings.Join(strings.Fields(string(k.Content)), " ")
	if len(args) > 0 && len(content) > args[0] {
		content = content[:args[0]] + "&hellip;"
	}
	p := bluemonday.StrictPolicy()
	return template.HTML(p.Sanitize(content))
}

func (k Kindy) MustTitle() string {
	if k.Type == KindyTypeLike || k.Type == KindyTypeRepost {
		url := k.LikeOf
		if k.Type == KindyTypeRepost {
			url = k.RepostOf
		}
		return k.Summary + " " + url
	}
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

	return string(k.Type)
}

func (k Kindy) MustDescription() string {
	if k.Summary != "" {
		return k.Summary
	}
	if k.Content != "" {
		p := bluemonday.StrictPolicy()
		return p.Sanitize(string(k.Content))
	}

	return string(k.Type)
}
