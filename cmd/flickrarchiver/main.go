package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
)

const (
	KindyDataPath    = "/kd/"
	KindyContentPath = "content/kindy/"
	KindyURLPhotos   = "/photos/"
)

type AutoGenerated struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	CountViews    string `json:"count_views"`
	CountFaves    string `json:"count_faves"`
	CountComments string `json:"count_comments"`
	DateTaken     string `json:"date_taken"`
	CountTags     string `json:"count_tags"`
	CountNotes    string `json:"count_notes"`
	Rotation      int    `json:"rotation"`
	DateImported  string `json:"date_imported"`
	Photopage     string `json:"photopage"`
	Original      string `json:"original"`
	License       string `json:"license"`
	Geo           []struct {
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
		Accuracy  string `json:"accuracy"`
	} `json:"geo"`
	Groups []any `json:"groups"`
	Albums []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"albums"`
	Tags []struct {
		Tag        string `json:"tag"`
		User       string `json:"user"`
		DateCreate string `json:"date_create"`
	} `json:"tags"`
	People             []any  `json:"people"`
	Notes              []any  `json:"notes"`
	Privacy            string `json:"privacy"`
	CommentPermissions string `json:"comment_permissions"`
	TaggingPermissions string `json:"tagging_permissions"`
	Safety             string `json:"safety"`
	Comments           []any  `json:"comments"`
}

var urlRe = regexp.MustCompile("[^a-zA-Z0-9- ]+")

func main() {
	pd := flag.String("photodir", "", "Directory storing your Flickr archive, merge all your photos into one folder")
	jd := flag.String("jsondir", "", "JSON dir containing your posts")
	flag.Parse()

	if *pd == "" || *jd == "" {
		log.Fatal("the flags `photodir`, `jsondir` are required")
	}

	// Find all photo JSON files
	photoFiles, err := findPhotoJSON(*jd)
	if err != nil {
		log.Fatalf("failed to find photo JSON files: %v", err)
	}
	var photos []AutoGenerated
	for _, file := range photoFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("failed to read photo JSON file: %v", err)
		}
		var photo AutoGenerated
		if err := json.Unmarshal(b, &photo); err != nil {
			log.Fatalf("failed to collect your photo: %v", err)
		}
		photos = append(photos, photo)
	}

	// Convert to Kindy JSON
	for _, v := range photos {
		publishedAt, err := time.Parse("2006-01-02 15:04:05", v.DateTaken)
		if err != nil {
			log.Printf("failed to parse date for %s: %v\n", v.DateTaken, err)
			continue
		}
		var geo *local.KindyGeo
		if len(v.Geo) > 0 && v.Geo[0].Latitude != "" && v.Geo[0].Longitude != "" {
			// add dot to latitude and longitude
			latstring, lonstring := addDotToGeoWith6Decimal(v.Geo[0].Latitude, v.Geo[0].Longitude)
			lat, _ := strconv.ParseFloat(latstring, 64)
			lng, _ := strconv.ParseFloat(lonstring, 64)
			geo = &local.KindyGeo{
				Latitude:  lat,
				Longitude: lng,
			}
		}
		var tags []string
		for _, tag := range v.Tags {
			tags = append(tags, tag.Tag)
		}

		url := constructURL(v.ID, v.Name)

		k, err := createKindyPhoto(url, v.Name, v.Description, v.Photopage, tags, publishedAt, geo)
		if err != nil {
			log.Printf("failed to create photo for %s: %v\n", url, err)
		}

		log.Printf("created photo: %v", k.Permalink)
	}
}

func createKindyPhoto(url, title, description, flickrUrl string, tags []string, publishedAt time.Time, geo *local.KindyGeo) (*local.Kindy, error) {
	if url == "" {
		return nil, errors.New("can't publish empty URL")
	}

	slug := fmt.Sprintf("%x", md5.Sum([]byte(publishedAt.Format(time.RFC3339))))

	entry := local.Kindy{
		Type:        local.KindyTypePhoto,
		Title:       title,
		Summary:     template.HTML(description),
		Content:     template.HTML(KindyDataPath + "photos/flickr/" + url),
		PublishedAt: publishedAt,
		Slug:        slug,
		Permalink:   KindyURLPhotos + slug,
		Geo:         geo,
		Tags:        tags,
		Syndication: []local.KindySyndication{{URL: flickrUrl, Type: "flickr"}},
	}

	b, err := json.MarshalIndent(entry, "", "    ")
	if err != nil {
		return nil, err
	}

	return &entry, os.WriteFile(KindyContentPath+entry.Permalink+".json", b, 0644)
}

func findPhotoJSON(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(filepath.Base(path), "photo_") && filepath.Ext(path) == ".json" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func addDotToGeoWith6Decimal(lat, long string) (string, string) {
	// reverse strings first
	lat = reverseString(lat)
	long = reverseString(long)
	latstring := lat[:6] + "." + lat[6:]
	longstring := long[:6] + "." + long[6:]
	// reverse strings back
	return reverseString(latstring), reverseString(longstring)
}

func reverseString(s string) string {
	var result string
	for i := len(s) - 1; i >= 0; i-- {
		result += string(s[i])
	}
	return result
}

func constructURL(id, name string) string {
	// id 9696863174
	// name Sint-Sebastiaansguild
	// result sint-sebastiaansguild_9696863174_o.jpg

	name = urlRe.ReplaceAllString(name, "")

	return fmt.Sprintf("%s_%s_o.jpg", strings.ToLower(strings.ReplaceAll(name, " ", "-")), id)
}
