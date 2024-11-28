package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	local "github.com/gerbenjacobs/gerben.dev"
)

func (h *Handler) Kindy(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/kindy.gohtml")...))

	b, err := os.ReadFile("content/kindy" + r.URL.Path + ".json")
	if err != nil {
		slog.Error("failed to read file", "file", r.URL.Path, "error", err)
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	var kind local.Kindy
	if err := json.Unmarshal(b, &kind); err != nil {
		slog.Error("failed to unmarshal kindy", "file", r.URL.Path, "error", err)
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	// tz, _ := time.LoadLocation("Europe/Amsterdam")
	// testKind := local.Kindy{
	// 	Type:   "note",
	// 	MFType: "h-entry",
	// 	// Title:       "Help, I have created a microformatted post",
	// 	Summary:     "Well, apart from webmentions working, this should be my first proper note! Hello world! 👋",
	// 	PublishedAt: time.Date(2024, 11, 28, 1, 43, 0, 0, tz),
	// 	// Content:     "Hello world, I'm read from a file!<br>Actually, that's a joke.. <strong>I'm really just a hardcoded string</strong><p>But I do have HTML!</p>",
	// 	Permalink: r.RequestURI,
	// 	// LikeOf:    "https://todon.nl/@gerben/113556315040488570",
	// 	Author: &local.KindyAuthor{
	// 		Name:  "Gerben Jacobs",
	// 		URL:   "https://gerben.dev",
	// 		Photo: "https://gerben.dev/images/avatar.jpg",
	// 	},
	// 	// Syndication: []local.KindySyndication{
	// 	// 	{
	// 	// 		Type: "twitter",
	// 	// 		URL:  "https://twitter.com/gerbenjacobs/6jh2467437",
	// 	// 	},
	// 	// 	{
	// 	// 		Type: "fediverse",
	// 	// 		URL:  "https://todon.nl/@gerben/326354347j432",
	// 	// 	},
	// 	// },
	// }
	// b, err = json.Marshal(testKind)
	// log.Print(string(b), err)

	if err := t.Execute(w, kind); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
