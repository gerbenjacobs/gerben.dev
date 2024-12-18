package handler

import (
	"html/template"
	"net/http"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) timeline(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/timeline.gohtml")...))

	entries := internal.GetTimelineData()
	author, _ := getAuthor()

	type pageData struct {
		Metadata internal.Metadata
		Author   *local.KindyAuthor
		Entries  []local.Kindy
	}
	data := pageData{
		Metadata: internal.Metadata{
			Title:       "Timeline",
			Description: "This page lists all notes, reposts and likes on gerben.dev in chronological order.",
		},
		Author:  author,
		Entries: entries,
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) timelineXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	http.ServeFile(w, r, internal.TimelineRssCache)
}
