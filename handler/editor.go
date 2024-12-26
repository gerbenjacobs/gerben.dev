package handler

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	kindy "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

const cookieName = "flash"

func kindyEditor(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/views/kindy/editor.html"))

	type kindyEditorStruct struct {
		Author kindy.KindyAuthor
		Entry  kindy.Kindy
		Tags   []string
		Flash  string
	}

	// handle POST
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			slog.Error("failed to parse POST form", "error", err)
			http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
			http.Redirect(w, r, kindy.KindyEditorPath, http.StatusFound)
			return
		}
		_ = r.ParseMultipartForm(10 << 20) // 10 MB

		if r.PostForm.Get("type") == "author" {
			if err := postAuthor(r.PostForm); err != nil {
				slog.Error("failed to store author", "error", err)
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
			}
			http.Redirect(w, r, kindy.KindyEditorPath, http.StatusFound)
			return
		}

		if r.PostForm.Get("type") == "note" {
			entry, err := postNote(r.PostForm)
			if err != nil {
				slog.Error("failed to publish note", "error", err)
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
				http.Redirect(w, r, kindy.KindyEditorPath, http.StatusFound)
				return
			}
			internal.CreateCaches()
			http.Redirect(w, r, kindy.KindyURLNotes+entry.Slug, http.StatusFound)
			return
		}

		if r.PostForm.Get("type") == "like" {
			entry, err := postLike(r.PostForm)
			if err != nil {
				slog.Error("failed to publish like", "error", err)
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
				http.Redirect(w, r, kindy.KindyEditorPath, http.StatusFound)
				return
			}
			internal.CreateCaches()
			http.Redirect(w, r, kindy.KindyURLLikes+entry.Slug, http.StatusFound)
			return
		}

		if r.PostForm.Get("type") == "repost" {
			entry, err := postRepost(r.PostForm)
			if err != nil {
				slog.Error("failed to publish repost", "error", err)
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
				http.Redirect(w, r, kindy.KindyEditorPath, http.StatusFound)
				return
			}
			internal.CreateCaches()
			http.Redirect(w, r, kindy.KindyURLReposts+entry.Slug, http.StatusFound)
			return
		}

		if r.PostForm.Get("type") == "photos" {
			entry, err := postPhotos(r)
			if err != nil {
				slog.Error("failed to publish photos", "error", err)
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
				http.Redirect(w, r, kindy.KindyEditorPath, http.StatusFound)
				return
			}
			internal.CreateCaches()
			http.Redirect(w, r, kindy.KindyURLPhotos+entry.Slug, http.StatusFound)
			return
		}

		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "Nothing was updated."})
		http.Redirect(w, r, kindy.KindyEditorPath, http.StatusFound)
		return
	}

	// handle GET
	data := kindyEditorStruct{
		Tags: internal.GetTags(),
	}

	// author for author form
	author, err := getAuthor()
	if err != nil {
		slog.Warn("kindy: failed to get author", "error", err)
	} else {
		data.Author = *author
	}

	// get flash message cookie
	c, err := r.Cookie(cookieName)
	if err != nil {
		slog.Warn("failed to get cookie", "error", err)
	} else {
		// store message
		data.Flash = c.Value
		// delete cookie
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: ""})
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func getAuthor() (*kindy.KindyAuthor, error) {
	var author kindy.KindyAuthor
	b, err := os.ReadFile(kindy.KindyContentPath + "author.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &author); err != nil {
		return nil, err
	}

	return &author, nil
}

func postAuthor(data url.Values) error {
	author := kindy.KindyAuthor{
		Name:  data.Get("name"),
		URL:   data.Get("url"),
		Photo: data.Get("photo"),
	}
	b, err := json.MarshalIndent(author, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(kindy.KindyContentPath+"author.json", b, 0644)
}

func postNote(data url.Values) (*kindy.Kindy, error) {
	if data.Get("content") == "" {
		return nil, errors.New("can't publish empty content")
	}

	publishedAt := time.Now()
	if data.Get("publishedat") != "" {
		pa, err := time.Parse("2006-01-02T15:04", data.Get("publishedat"))
		if err != nil {
			return nil, err
		}
		if !pa.IsZero() {
			publishedAt = pa
		}
	}
	slug := fmt.Sprintf("%x", md5.Sum([]byte(publishedAt.Format(time.RFC3339))))

	if data.Get("slug") != "" {
		slug = getTitleURLFromString(data.Get("slug"))
	}

	// handle tags
	var tags []string
	for _, t := range strings.Split(data.Get("tags"), ",") {
		tt := strings.TrimSpace(t)
		if tt != "" {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	entry := kindy.Kindy{
		Type:        kindy.KindyTypeNote,
		PublishedAt: publishedAt,
		Slug:        slug,
		Permalink:   kindy.KindyURLNotes + slug,
		Tags:        tags,
	}

	// handle markdown
	if data.Has("markdown") {
		entry.Markdown = data.Get("content")
	} else {
		entry.Content = template.HTML(data.Get("content"))
	}

	b, err := json.MarshalIndent(entry, "", "    ")
	if err != nil {
		return nil, err
	}

	return &entry, os.WriteFile(kindy.KindyContentPath+entry.Permalink+".json", b, 0644)
}

func postLike(data url.Values) (*kindy.Kindy, error) {
	if data.Get("url") == "" {
		return nil, errors.New("can't publish empty URL")
	}

	publishedAt := time.Now()
	slug := fmt.Sprintf("%x", md5.Sum([]byte(publishedAt.Format(time.RFC3339))))

	entry := kindy.Kindy{
		Type:        kindy.KindyTypeLike,
		Summary:     kindy.KindySummaryLike,
		LikeOf:      data.Get("url"),
		PublishedAt: publishedAt,
		Slug:        slug,
		Permalink:   kindy.KindyURLLikes + slug,
	}

	b, err := json.MarshalIndent(entry, "", "    ")
	if err != nil {
		return nil, err
	}

	return &entry, os.WriteFile(kindy.KindyContentPath+entry.Permalink+".json", b, 0644)
}

func postRepost(data url.Values) (*kindy.Kindy, error) {
	if data.Get("url") == "" {
		return nil, errors.New("can't publish empty URL")
	}

	publishedAt := time.Now()
	slug := fmt.Sprintf("%x", md5.Sum([]byte(publishedAt.Format(time.RFC3339))))

	entry := kindy.Kindy{
		Type:        kindy.KindyTypeRepost,
		Summary:     kindy.KindySummaryRepost,
		RepostOf:    data.Get("url"),
		PublishedAt: publishedAt,
		Slug:        slug,
		Permalink:   kindy.KindyURLReposts + slug,
	}

	b, err := json.MarshalIndent(entry, "", "    ")
	if err != nil {
		return nil, err
	}

	return &entry, os.WriteFile(kindy.KindyContentPath+entry.Permalink+".json", b, 0644)
}

func postPhotos(req *http.Request) (*kindy.Kindy, error) {
	data := req.PostForm

	// handle photo
	in, header, err := req.FormFile("photo")
	if err != nil {
		return nil, err
	}
	defer in.Close()

	f, err := internal.HandleUploadedFile(in, header)
	if err != nil {
		return nil, err
	}

	publishedAt := time.Now()
	if data.Get("publishedat") != "" {
		pa, err := time.Parse("2006-01-02T15:04", data.Get("publishedat"))
		if err != nil {
			return nil, err
		}
		if !pa.IsZero() {
			publishedAt = pa
		}
	}
	slug := fmt.Sprintf("%x", md5.Sum([]byte(publishedAt.Format(time.RFC3339))))

	if data.Get("slug") != "" {
		slug = getTitleURLFromString(data.Get("slug"))
	}

	// handle tags
	var tags []string
	for _, t := range strings.Split(data.Get("tags"), ",") {
		tt := strings.TrimSpace(t)
		if tt != "" {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	entry := kindy.Kindy{
		Type:        kindy.KindyTypePhoto,
		Title:       data.Get("title"),
		Summary:     template.HTML(data.Get("summary")),
		Content:     template.HTML(kindy.KindyDataPath + "photos/" + filepath.Base(f.Name())),
		PublishedAt: publishedAt,
		Slug:        slug,
		Permalink:   kindy.KindyURLPhotos + slug,
		Tags:        tags,
	}

	b, err := json.MarshalIndent(entry, "", "    ")
	if err != nil {
		return nil, err
	}

	return &entry, os.WriteFile(kindy.KindyContentPath+entry.Permalink+".json", b, 0644)
}

func getTitleURLFromString(title string) (output string) {
	// first, strip out any special characters
	re := regexp.MustCompile(`(?m)[^\d^A-Z^a-z^\-^\s]`)
	substitution := ""
	output = re.ReplaceAllString(title, substitution)

	// set to lowercase
	output = strings.ToLower(output)

	// next, replace all whitespace characters with dashes
	re = regexp.MustCompile(`(?m)[\s]`)
	substitution = "-"
	output = re.ReplaceAllString(output, substitution)

	// replace "clumps" of 2 or more hyphens with 1 hyphen
	re = regexp.MustCompile(`(?m)-{2,}`)
	substitution = "-"
	output = re.ReplaceAllString(output, substitution)

	// result is only up to 36 characters (or the whole thing if less than 36)
	output = output[:int(math.Min(float64(len(output)), 36))]

	// remove trailing hyphens from the final output
	re = regexp.MustCompile(`(?m)-*$`)
	substitution = ""
	output = re.ReplaceAllString(output, substitution)

	return output
}
