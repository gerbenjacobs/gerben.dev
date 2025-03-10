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
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/timeline.gohtml", "static/views/partials/timeline-paginated.gohtml")...))

	// handle content type queries
	showNotes := r.URL.Query().Get("showNotes") == "on"
	showReplies := r.URL.Query().Get("showReplies") == "on"
	showReposts := r.URL.Query().Get("showReposts") == "on"
	showLikes := r.URL.Query().Get("showLikes") == "on"

	if !r.URL.Query().Has("showNotes") && !r.URL.Query().Has("showReplies") && !r.URL.Query().Has("showReposts") && !r.URL.Query().Has("showLikes") {
		// nothing selected, default to show all
		showNotes = true
		showReplies = true
		showReposts = true
		showLikes = true
	}

	// handle since query
	oldestContentDate := time.Date(2024, 11, 28, 0, 0, 0, 0, time.UTC)
	since, upto, cursor := h.handleTimePagination(r.URL.Query().Get("since"), oldestContentDate)

	// get data
	entries := internal.GetTimelineData(since, &upto, showNotes, showReplies, showReposts, showLikes)

	// if last entry is 'younger' than upto value, we have reached the end
	if len(entries) > 0 && entries[len(entries)-1].PublishedAt.Before(upto) {
		cursor = ""
	}

	author, _ := getAuthor()
	type pageData struct {
		Metadata    internal.Metadata
		Author      *local.KindyAuthor
		Entries     []local.Kindy
		NewSince    string
		ShowNotes   bool
		ShowReplies bool
		ShowReposts bool
		ShowLikes   bool
	}
	data := pageData{
		Metadata: internal.Metadata{
			Env:         Env,
			Title:       "Timeline",
			Description: "This page lists all notes, reposts and likes on gerben.dev in chronological order.",
			Permalink:   "/timeline",
		},
		Author:      author,
		Entries:     entries,
		NewSince:    cursor,
		ShowNotes:   showNotes,
		ShowReplies: showReplies,
		ShowReposts: showReposts,
		ShowLikes:   showLikes,
	}

	if len(entries) == 0 {
		http.Error(w, "no entries found", http.StatusNotFound)
		return
	}

	// if HTMX call, we return partials only
	isHX := r.Header.Get("HX-Request") //r.URL.Query().Get("HX-Request") to test
	if isHX == "true" {
		if err := t.ExecuteTemplate(w, "timeline-paginated", data); err != nil {
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

func (h *Handler) handleTimePagination(timeString string, lastTime time.Time) (from, to time.Time, cursor string) {
	from = time.Now()
	sinceQuery := timeString
	if sinceQuery != "" {
		s, err := time.Parse("2006-01-02", sinceQuery)
		if err != nil {
			slog.Error("failed to parse since query", "error", err)
		} else {
			from = s
		}
	}
	to = from.AddDate(0, -1, -7)
	cursor = to.Format("2006-01-02")
	if to.Before(lastTime) {
		// stop loading if we're going back too far
		cursor = ""
	}

	return
}
