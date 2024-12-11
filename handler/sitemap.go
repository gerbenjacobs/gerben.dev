package handler

import (
	"bytes"
	"encoding/gob"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) sitemap(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/sitemap.html")...))

	cacheFile := ".cache/sitemap.gob"
	b, err := internal.GetCache(cacheFile, 24*time.Hour)
	slog.Warn("checking cache", "error", err)
	if err != nil {
		posts, err := GetKindyByType("posts")
		if err != nil {
			slog.Error("failed to load kindy posts", "error", err)
		}
		photos, err := GetKindyByType("photos")
		if err != nil {
			slog.Error("failed to load kindy photos", "error", err)
		}
		notes, err := GetKindyByType("notes")
		if err != nil {
			slog.Error("failed to load kindy notes", "error", err)
		}
		likes, err := GetKindyByType("likes")
		if err != nil {
			slog.Error("failed to load kindy likes", "error", err)
		}
		reposts, err := GetKindyByType("reposts")
		if err != nil {
			slog.Error("failed to load kindy reposts", "error", err)
		}
		// replies, err := GetKindyByType("replies")
		// if err != nil {
		// 	slog.Error("failed to load kindy replies", "error", err)
		// }
		gob.Register(local.Kindy{})
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(posts)
		enc.Encode(photos)
		enc.Encode(notes)
		enc.Encode(likes)
		enc.Encode(reposts)
		internal.SetCache(cacheFile, buf.Bytes())
		b = buf.Bytes()
	}

	var posts, photos, notes, likes, reposts []local.Kindy
	dec := gob.NewDecoder(bytes.NewReader(b))
	dec.Decode(&posts)
	dec.Decode(&photos)
	dec.Decode(&notes)
	dec.Decode(&likes)
	dec.Decode(&reposts)

	type pageData struct {
		Metadata internal.Metadata
		Counts   map[string]int
		Posts    []local.Kindy
		Photos   []local.Kindy
		Notes    []local.Kindy
		Likes    []local.Kindy
		Reposts  []local.Kindy
		Replies  []local.Kindy
	}

	data := pageData{
		Metadata: internal.Metadata{Title: "Sitemap", Description: "A HTML version of my sitemap"},
		Counts: map[string]int{
			"posts":   len(posts),
			"photos":  len(photos),
			"notes":   len(notes),
			"likes":   len(likes),
			"reposts": len(reposts),
			// "replies": len(replies),
		},
		Posts:   kindyLimit(posts, 10),
		Photos:  kindyLimit(photos, 10),
		Notes:   kindyLimit(notes, 10),
		Likes:   kindyLimit(likes, 10),
		Reposts: kindyLimit(reposts, 10),
		// Replies: kindyLimit(replies, 10),
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func kindyLimit(entries []local.Kindy, limit int) []local.Kindy {
	if len(entries) > limit {
		return entries[:limit]
	}

	return entries
}
