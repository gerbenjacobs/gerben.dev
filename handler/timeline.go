package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) timeline(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/timeline.gohtml", "static/views/partials/timeline-partial.gohtml")...))

	// handle since query
	since := time.Now()
	sinceQuery := r.URL.Query().Get("since")
	if sinceQuery != "" {
		s, err := time.Parse("2006-01-02", sinceQuery)
		if err != nil {
			slog.Error("failed to parse since query", "error", err)
		} else {
			since = s
		}
		slog.Info("loading timeline", "since", since.Format("2006-01-02"), "isHX", r.Header.Get("HX-Request"))
	}
	upto := since.AddDate(0, 0, -7)
	newSince := upto.Format("2006-01-02")
	if upto.Before(time.Date(2024, 11, 28, 0, 0, 0, 0, time.UTC)) {
		// stop loading if we're going back too far
		// first content: /likes/5a4c8aa4b77118632b80e5a2c85a1123 (2024-11-28)
		newSince = ""
	}

	// get data
	entries := internal.GetTimelineData(since, &upto)
	author, _ := getAuthor()
	type pageData struct {
		Metadata internal.Metadata
		Author   *local.KindyAuthor
		Entries  []local.Kindy
		NewSince string
	}
	data := pageData{
		Metadata: internal.Metadata{
			Title:       "Timeline",
			Description: "This page lists all notes, reposts and likes on gerben.dev in chronological order.",
		},
		Author:   author,
		Entries:  entries,
		NewSince: newSince,
	}

	if len(entries) == 0 {
		http.Error(w, "no entries found", http.StatusNotFound)
		return
	}

	// if HTMX call, we return partials only
	isHX := r.Header.Get("HX-Request") //r.URL.Query().Get("HX-Request") to test
	if isHX == "true" {
		if err := t.ExecuteTemplate(w, "timeline-partial", data); err != nil {
			http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) timelineXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	http.ServeFile(w, r, internal.TimelineRssCache)
}
