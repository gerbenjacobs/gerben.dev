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
	"regexp"
	"strings"
	"time"

	kindy "github.com/gerbenjacobs/gerben.dev"
)

const cookieName = "flash"

var (
	KindyContentPath = "content/kindy/"
	KindyBaseURL     = "https://gerben.dev/"
	KindyURLNotes    = "notes/"
)

func kindyEditor(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/views/kindy/editor.html"))

	type kindyEditorStruct struct {
		Author kindy.KindyAuthor
		Entry  kindy.Kindy
		Flash  string
	}

	// handle POST
	if r.Method == "POST" {
		r.ParseForm()

		if r.PostForm.Get("type") == "author" {
			if err := postAuthor(r.PostForm); err != nil {
				slog.Error("failed to store author", "error", err)
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
			}
			http.Redirect(w, r, "/kindy", http.StatusFound)
			return
		}

		if r.PostForm.Get("type") == "note" {
			note, err := postNote(r.PostForm)
			if err != nil {
				slog.Error("failed to publish note", "error", err)
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: err.Error()})
				http.Redirect(w, r, "/kindy", http.StatusFound)
				return
			}
			http.Redirect(w, r, KindyURLNotes+note.Slug, http.StatusFound)
			return
		}
	}

	// handle GET
	var data kindyEditorStruct

	// try to get author
	author, err := getAuthor()
	if err != nil {
		slog.Warn("kindy: failed to get author", "error", err)
	} else {
		data.Author = *author
	}

	// get flash message cookie
	c, err := r.Cookie("flash")
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
	b, err := os.ReadFile(KindyContentPath + "author.json")
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
	return os.WriteFile(KindyContentPath+"author.json", b, 0644)
}

func postNote(data url.Values) (*kindy.Kindy, error) {
	if data.Get("content") == "" {
		return nil, errors.New("can't publish empty content")
	}

	publishedAt := time.Now()
	slug := fmt.Sprintf("%x", md5.Sum([]byte(publishedAt.Format(time.RFC3339))))

	if data.Get("slug") != "" {
		slug = getTitleURLFromString(data.Get("slug"))
	}

	author, _ := getAuthor()
	note := kindy.Kindy{
		Type:        "note",
		MFType:      "h-entry",
		Summary:     data.Get("content"),
		Content:     template.HTML(data.Get("content")),
		PublishedAt: publishedAt,
		Slug:        slug,
		Permalink:   KindyBaseURL + KindyURLNotes + slug,
		Author:      author,
	}

	b, err := json.MarshalIndent(note, "", "    ")
	if err != nil {
		return nil, err
	}

	return &note, os.WriteFile(KindyContentPath+"notes/"+slug+".json", b, 0644)
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
