package internal

import (
	"encoding/xml"
	"log/slog"
	"slices"
	"sort"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
)

var MinimumTimelineEntries = 20

func CreateTimelineXML() ([]byte, error) {
	timelineCutoffDate := time.Now().AddDate(0, -3, 0)
	entries := GetTimelineData(time.Now(), &timelineCutoffDate, true, true, true, true)
	if len(entries) == 0 {
		return nil, nil
	}

	lastUpdatedTime := entries[0].PublishedAt
	feed := RssFeed{
		Title:       "@gerben.dev timeline",
		Link:        "https://gerben.dev/timeline",
		Description: "A collection of all my notes, reposts and likes, sorted by date.",
		PubDate:     lastUpdatedTime.Format(time.RFC1123Z),
	}

	for _, entry := range entries {
		url := "https://gerben.dev" + entry.Permalink
		linkUrl := url
		if entry.Type == local.KindyTypeRepost {
			linkUrl = entry.RepostOf
		}
		if entry.Type == local.KindyTypeLike {
			linkUrl = entry.LikeOf
		}
		item := &RssItem{
			Title:       Titlify(entry.MustTitle()),
			Link:        linkUrl,
			Guid:        &RssGuid{Id: url, IsPermaLink: "true"},
			Description: string(entry.MustDescription()),
			PubDate:     entry.PublishedAt.Format(time.RFC1123Z),
			Category:    entry.Tags,
		}
		if entry.GetContent() != "" {
			item.Content = &RssContent{Content: string(entry.GetContent())}
		}
		feed.Items = append(feed.Items, item)
	}

	fullRss := RssFeedXml{
		Version:          "2.0",
		ContentNamespace: "http://purl.org/rss/1.0/modules/content/",
		Channel:          &feed,
	}

	return xml.Marshal(fullRss)
}

func GetTimelineData(since time.Time, upto *time.Time, showNotes, showReplies, showReposts, showLikes bool) []local.Kindy {
	var entries []local.Kindy

	if showNotes {
		notes, _ := GetKindyCacheByType(local.KindyTypeNote)
		entries = append(entries, notes...)
	}
	if showReplies {
		replies, _ := GetKindyCacheByType(local.KindyTypeReplies)
		entries = append(entries, replies...)
	}
	if showReposts {
		reposts, _ := GetKindyCacheByType(local.KindyTypeRepost)
		entries = append(entries, reposts...)
	}
	if showLikes {
		likes, _ := GetKindyCacheByType(local.KindyTypeLike)
		entries = append(entries, likes...)
	}

	// Sort the entries on published date
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PublishedAt.After(entries[j].PublishedAt)
	})

	// Filter out entries if a since query is provided
	entries = slices.DeleteFunc(entries, func(e local.Kindy) bool {
		return e.PublishedAt.After(since)
	})

	// Filter out entries if an upto query is provided
	if len(entries) > MinimumTimelineEntries && upto != nil {
		tmpEntries := slices.Clone(entries)
		entries = slices.DeleteFunc(entries, func(e local.Kindy) bool {
			return e.PublishedAt.Before(*upto)
		})
		if len(entries) < MinimumTimelineEntries {
			slog.Warn("restoring entries to meet minimum", "current", len(entries), "minimum", MinimumTimelineEntries, "restoredFrom", len(tmpEntries))
			entries = tmpEntries[:MinimumTimelineEntries]
		}
	}

	return entries
}
