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
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
)

const (
	KindyDataPath    = "/kd/"
	KindyContentPath = "content/kindy/"
	KindyURLPhotos   = "/photos/"
)

type AutoGenerated struct {
	Media []struct {
		URI               string `json:"uri"`
		CreationTimestamp int    `json:"creation_timestamp"`
		MediaMetadata     struct {
			PhotoMetadata struct {
				ExifData []struct {
					Latitude         float64 `json:"latitude,omitempty"`
					Longitude        float64 `json:"longitude,omitempty"`
					SceneCaptureType string  `json:"scene_capture_type,omitempty"`
					Software         string  `json:"software,omitempty"`
					DeviceID         string  `json:"device_id,omitempty"`
					DateTimeOriginal string  `json:"date_time_original,omitempty"`
					SourceType       string  `json:"source_type,omitempty"`
				} `json:"exif_data"`
			} `json:"photo_metadata"`
			VideoMetadata struct {
				ExifData []struct {
					Latitude  float64 `json:"latitude,omitempty"`
					Longitude float64 `json:"longitude,omitempty"`
				} `json:"exif_data"`
			} `json:"video_metadata"`
			CameraMetadata struct {
				HasCameraMetadata bool `json:"has_camera_metadata"`
			} `json:"camera_metadata"`
		} `json:"media_metadata"`
		Title           string `json:"title"`
		CrossPostSource struct {
			SourceApp string `json:"source_app"`
		} `json:"cross_post_source"`
		BackupURI string `json:"backup_uri"`
	} `json:"media"`
}

var hashtagRe = regexp.MustCompile(`[^#]*?#(\w+)`)
var hashtagRemoverRe = regexp.MustCompile(`#(\w+)`)

func main() {
	wd := flag.String("workdir", "", "Working directory of your Instagram archive")
	od := flag.String("outputdir", "", "Output directory i.e. your Kindy content folder")
	pf := flag.String("postsfile", "/your_instagram_activity/content/posts_1.json", "JSON file containing your posts, as seen from workdir")
	flag.Parse()

	if *wd == "" || *od == "" {
		log.Fatalf("the flags `workdir` and `outputdir` are required\nworkdir: %#v\noutputdir: %#v\n", *wd, *od)
	}

	postsFile := filepath.Join(*wd + *pf)
	b, err := os.ReadFile(postsFile)
	if err != nil {
		log.Fatalf("failed to read your posts file: %v", err)
	}

	var posts []AutoGenerated
	if err := json.Unmarshal(b, &posts); err != nil {
		log.Fatalf("failed to collect your posts: %v", err)
	}

	for _, v := range posts {
		publishedAt := time.Unix(int64(v.Media[0].CreationTimestamp), 0)
		var geo *local.KindyGeo
		for _, v := range v.Media[0].MediaMetadata.PhotoMetadata.ExifData {
			if v.Latitude != 0 && v.Longitude != 0 {
				geo = &local.KindyGeo{
					Latitude:  v.Latitude,
					Longitude: v.Longitude,
				}
			}
		}
		if geo == nil {
			for _, v := range v.Media[0].MediaMetadata.VideoMetadata.ExifData {
				if v.Latitude != 0 && v.Longitude != 0 {
					geo = &local.KindyGeo{
						Latitude:  v.Latitude,
						Longitude: v.Longitude,
					}
				}
			}
		}

		k, err := createKindyPhoto(v.Media[0].URI, v.Media[0].Title, publishedAt, geo)
		if err != nil {
			log.Printf("failed to create photo for %s: %v\n", v.Media[0].URI, err)
		}

		log.Printf("created photo: %v", k.Permalink)
	}
}

func createKindyPhoto(url, title string, publishedAt time.Time, geo *local.KindyGeo) (*local.Kindy, error) {
	if url == "" {
		return nil, errors.New("can't publish empty URL")
	}

	slug := fmt.Sprintf("%x", md5.Sum([]byte(publishedAt.Format(time.RFC3339))))

	var tags []string
	matches := hashtagRe.FindAllStringSubmatch(title, -1)
	for _, match := range matches {
		tags = append(tags, match[1])
	}
	title = hashtagRemoverRe.ReplaceAllString(title, "")

	author, _ := getAuthor()
	entry := local.Kindy{
		Type:        local.KindyTypePhoto,
		Summary:     title,
		Content:     template.HTML(KindyDataPath + "photos/" + url),
		PublishedAt: publishedAt,
		Slug:        slug,
		Permalink:   KindyURLPhotos + slug,
		Author:      author,
		Geo:         geo,
		Tags:        tags,
	}

	b, err := json.MarshalIndent(entry, "", "    ")
	if err != nil {
		return nil, err
	}

	return &entry, os.WriteFile(KindyContentPath+entry.Permalink+".json", b, 0644)
}

func getAuthor() (*local.KindyAuthor, error) {
	var author local.KindyAuthor
	b, err := os.ReadFile(KindyContentPath + "author.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &author); err != nil {
		return nil, err
	}

	return &author, nil
}
