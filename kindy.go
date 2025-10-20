package gerbendev

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
	"go.abhg.dev/goldmark/hashtag"
)

var gm goldmark.Markdown

type KindyType string

const (
	KindyEditorPath  = "/kindy"
	KindyDataPath    = "/kd/"
	KindyContentPath = "content/kindy/"

	KindyURLLikes   = "/likes/"
	KindyURLNotes   = "/notes/"
	KindyURLPhotos  = "/photos/"
	KindyURLPosts   = "/posts/"
	KindyURLReposts = "/reposts/"
	KindyURLReplies = "/replies/"

	KindySummaryLike   = "Liked"
	KindySummaryRepost = "Reposted"

	KindyTypeNote    KindyType = "note"
	KindyTypePost    KindyType = "post"
	KindyTypePhoto   KindyType = "photo"
	KindyTypeLike    KindyType = "like"
	KindyTypeRepost  KindyType = "repost"
	KindyTypeReplies KindyType = "reply"
)

type HashtagResolver struct{}

func (r *HashtagResolver) ResolveHashtag(n *hashtag.Node) (destination []byte, err error) {
	// if parent type is a link, we don't want to resolve it
	if n.Parent().Kind() == ast.KindLink {
		return nil, nil
	}
	return append([]byte("/tags/"), n.Tag...), nil
}

func init() {
	gm = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&frontmatter.Extender{},
			&hashtag.Extender{
				Resolver: &HashtagResolver{},
				Variant:  hashtag.ObsidianVariant,
				Attributes: []hashtag.Attribute{
					{
						Name:  "class",
						Value: "p-category",
					},
				},
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

// Kindy is a datastructure for content that adheres to Microformats 2
type Kindy struct {
	Type        KindyType          `json:"type"`
	Title       string             `json:"title,omitempty"`
	Summary     template.HTML      `json:"summary,omitempty"`
	Content     template.HTML      `json:"content,omitempty"`
	Markdown    string             `json:"markdown,omitempty"`
	PublishedAt time.Time          `json:"publishedAt"`
	Slug        string             `json:"slug,omitempty"`
	Permalink   string             `json:"permalink,omitempty"`
	Author      *KindyAuthor       `json:"author,omitempty"`
	Syndication []KindySyndication `json:"syndication,omitempty"`
	LikeOf      string             `json:"likeOf,omitempty"`
	RepostOf    string             `json:"repostOf,omitempty"`
	ReplyTo     string             `json:"replyTo,omitempty"`
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

func (k Kindy) Thumbnail() string {
	if k.Type == KindyTypePhoto {
		filePath := string(k.Content)
		ext := filepath.Ext(filePath)
		return fmt.Sprintf("%s_thumb%s", strings.TrimSuffix(filePath, ext), ext)
	}

	return ""
}

// HasContent returns true if the Kindy has content
// either as Markdown or as HTML
func (k Kindy) HasContent() bool {
	return k.Markdown != "" || k.Content != ""
}

// GetContent returns content or Markdown if it exists
func (k Kindy) GetContent() template.HTML {
	if k.Markdown != "" {
		return template.HTML(MarkdownToHTML(k.Markdown))
	}
	return k.Content
}

// ContentStripped strips all HTML with a strict policy
// but still returns a template.HTML so that properly escaped HTML entities still work
// It has an 'optional' args list, but we really only except 1 int which limits the length
func (k Kindy) ContentStripped(args ...int) template.HTML {
	p := bluemonday.StrictPolicy()
	content := p.Sanitize(string(k.GetContent()))
	content = strings.Join(strings.Fields(content), " ")
	if len(args) > 0 && len(content) > args[0] {
		content = content[:args[0]] + "&hellip;"
	}
	return template.HTML(content)
}

func (k Kindy) MustTitle() string {
	if k.Type == KindyTypeLike || k.Type == KindyTypeRepost {
		url := k.LikeOf
		if k.Type == KindyTypeRepost {
			url = k.RepostOf
		}
		return string(k.Summary) + " " + url
	}
	if k.Title != "" {
		return k.Title
	}
	if k.Summary != "" {
		return string(k.Summary)
	}
	if k.Type == KindyTypeNote || k.Type == KindyTypeReplies {
		// for Notes, we might as well use the content
		p := bluemonday.StrictPolicy()
		return p.Sanitize(string(k.GetContent()))
	}
	if k.Permalink != "" {
		return k.Permalink
	}

	return string(k.Type)
}

func (k Kindy) MustDescription() template.HTML {
	p := bluemonday.StrictPolicy()
	if k.Summary != "" {
		return template.HTML(p.Sanitize(string(k.Summary)))
	}
	if k.Title != "" {
		return template.HTML(p.Sanitize(string(k.Title)))
	}
	content := k.GetContent()
	if content != "" {
		return template.HTML(p.Sanitize(string(content)))
	}

	return template.HTML(k.MustTitle())
}

func (k Kindy) HasFlickrSyndication() bool {
	for _, s := range k.Syndication {
		if s.Type == "flickr" {
			return true
		}
	}
	return false
}

func (k Kindy) TimeAgo() string {
	return humanize.Time(k.PublishedAt)
}

func (kt KindyType) Emoji() string {
	emojis := map[KindyType]string{
		KindyTypePost:    "📝",
		KindyTypeNote:    "📜",
		KindyTypePhoto:   "📸",
		KindyTypeRepost:  "🔁",
		KindyTypeLike:    "⭐",
		KindyTypeReplies: "💬",
	}
	if v, ok := emojis[kt]; ok {
		return v
	}
	// default to post
	return emojis[KindyTypePost]
}

func (kt KindyType) URL() string {
	switch kt {
	case KindyTypeNote:
		return KindyURLNotes
	case KindyTypePhoto:
		return KindyURLPhotos
	case KindyTypePost:
		return KindyURLPosts
	case KindyTypeRepost:
		return KindyURLReposts
	case KindyTypeLike:
		return KindyURLLikes
	case KindyTypeReplies:
		return KindyURLReplies
	default:
		return ""
	}
}

func MarkdownToHTML(md string) string {
	var buf bytes.Buffer
	if err := gm.Convert([]byte(md), &buf); err != nil {
		slog.Error("failed to convert markdown to html", "error", err)
	}

	return buf.String()
}
