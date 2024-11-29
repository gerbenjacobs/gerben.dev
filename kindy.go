package gerbendev

import (
	"html/template"
	"time"
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
