package handler

import (
	"net/http"

	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) timelineXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	http.ServeFile(w, r, internal.TimelineRssCache)
}

func (h *Handler) photosXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	http.ServeFile(w, r, internal.PhotosRssCache)
}

func (h *Handler) postsXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	http.ServeFile(w, r, internal.PostsRssCache)
}
