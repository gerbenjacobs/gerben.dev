package internal

import (
	"encoding/xml"
	"html/template"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
)

func GetPostContent(content string) (template.HTML, error) {
	// for HTML content
	if strings.HasSuffix(content, ".html") {
		b, err := os.ReadFile("content/kindy" + content)
		if err != nil {
			return template.HTML(content), err
		}
		return template.HTML(b), nil
	}

	// for markdown content
	if strings.HasSuffix(content, ".md") {
		b, err := os.ReadFile("content/kindy" + content)
		if err != nil {
			return template.HTML(content), err
		}
		return template.HTML(local.MarkdownToHTML(string(b))), nil
	}
	return template.HTML(content), nil
}

func CreatePostsXML() ([]byte, error) {
	timelineCutoffDate := time.Now().AddDate(-1, 0, 0)
	entries, err := GetKindyCacheByType(local.KindyTypePost)
	if err != nil {
		return nil, err
	}
	entries = slices.DeleteFunc(entries, func(e local.Kindy) bool {
		return e.PublishedAt.Before(timelineCutoffDate)
	})

	lastUpdatedTime := entries[0].PublishedAt
	feed := RssFeed{
		Title:         "gerben.dev posts",
		Link:          "https://gerben.dev/posts/",
		Description:   "The latest blog posts on gerben.dev.",
		PubDate:       lastUpdatedTime.Format(time.RFC1123Z),
		LastBuildDate: time.Now().Format(time.RFC1123Z),
	}

	for _, entry := range entries {
		url := "https://gerben.dev" + entry.Permalink
		item := &RssItem{
			Title:       Titlify(entry.MustTitle()),
			Link:        url,
			Guid:        &RssGuid{Id: url, IsPermaLink: "true"},
			Description: string(entry.MustDescription()),
			PubDate:     entry.PublishedAt.Format(time.RFC1123Z),
			Category:    entry.Tags,
		}
		if entry.GetContent() != "" {
			content, err := GetPostContent(string(entry.Content))
			if err != nil {
				slog.Error("failed to get post content for posts.xml", "file", entry.Permalink, "error", err)
				continue
			}
			item.Content = &RssContent{Content: string(content)}
		}
		feed.Items = append(feed.Items, item)
	}

	fullRss := RssFeedXml{
		Version:          "2.0",
		ContentNamespace: "http://purl.org/rss/1.0/modules/content/",
		Channel:          &feed,
	}

	return xml.MarshalIndent(fullRss, "", "  ")
}
