package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) sitemap(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/sitemap.html")...))

	kindyData := map[local.KindyType][]local.Kindy{}
	for _, kindyType := range internal.KindyTypes {
		entries, err := internal.GetKindyCacheByType(kindyType)
		if err != nil {
			slog.Error("failed to load entries", "type", kindyType, "error", err)
			http.Error(w, "failed to load entries: "+err.Error(), http.StatusInternalServerError)
			return
		}
		kindyData[kindyType] = entries
	}

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
			"posts":   len(kindyData[local.KindyTypePost]),
			"photos":  len(kindyData[local.KindyTypePhoto]),
			"notes":   len(kindyData[local.KindyTypeNote]),
			"likes":   len(kindyData[local.KindyTypeLike]),
			"reposts": len(kindyData[local.KindyTypeRepost]),
		},
		Posts:   kindyLimit(kindyData[local.KindyTypePost], 10),
		Photos:  kindyLimit(kindyData[local.KindyTypePhoto], 10),
		Notes:   kindyLimit(kindyData[local.KindyTypeNote], 10),
		Likes:   kindyLimit(kindyData[local.KindyTypeLike], 10),
		Reposts: kindyLimit(kindyData[local.KindyTypeRepost], 10),
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
