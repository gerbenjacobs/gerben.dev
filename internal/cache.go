package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
)

var (
	ErrCacheExpired = errors.New("cache expired")
	ErrCacheCreated = errors.New("cache created")
)

// KindyTypes are all the active Kindy types on this site
var KindyTypes = []local.KindyType{local.KindyTypeLike, local.KindyTypePhoto, local.KindyTypePost, local.KindyTypeRepost, local.KindyTypeNote}

var (
	KindyTagsCache    = ".cache/kindy_tags.json"
	KindyPostsCache   = ".cache/kindy_posts.json"
	KindyPhotosCache  = ".cache/kindy_photos.json"
	KindyRepostsCache = ".cache/kindy_reposts.json"
	KindyLikesCache   = ".cache/kindy_likes.json"
	KindyNotesCache   = ".cache/kindy_notes.json"
)

func GetCache(filePath string, expiry time.Duration) ([]byte, error) {
	info, err := os.Stat(filePath)
	switch {
	// if no expiry, skip switch, always fetch from cache
	case expiry == 0:
		break
	case os.IsNotExist(err):
		_, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		return nil, ErrCacheCreated
	case err != nil:
		return nil, err
	case info.ModTime().Before(time.Now().Add(-10 * expiry)):
		return nil, ErrCacheExpired
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func SetCache(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

// CreateCaches creates the caches all data, on startup
func CreateCaches() error {
	// kindy caches
	for _, kindyType := range KindyTypes {
		if err := CreateKindyCacheByType(kindyType); err != nil {
			return err
		}
	}

	// tag cache
	if err := CreateTagCache(); err != nil {
		return err
	}

	return nil
}

func GetKindyCacheByType(kind local.KindyType) ([]local.Kindy, error) {
	cacheFile := kindyCacheFile(kind)
	b, err := GetCache(cacheFile, 0)
	if err != nil {
		slog.Warn("failed to load cache", "error", err)
		return nil, err
	}

	if len(b) <= 2 {
		return nil, nil
	}

	var entries []local.Kindy
	if err := json.Unmarshal(b, &entries); err != nil {
		slog.Error("failed to unmarshal cache", "error", err)
		return nil, err
	}

	return entries, nil
}

func CreateKindyCacheByType(kind local.KindyType) error {
	entries, err := GetKindyByType(kind)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	cacheFile := kindyCacheFile(kind)
	return SetCache(cacheFile, b)
}

func CreateTagCache() error {
	var entries []local.Kindy
	for _, kindyType := range KindyTypes {
		kinds, _ := GetKindyCacheByType(kindyType)
		entries = append(entries, kinds...)
	}

	tags := map[string]TagInfo{}
	for _, entry := range entries {
		fp := fmt.Sprintf("%s%s/%s.json", local.KindyContentPath, entry.Type+"s", entry.Slug)
		mergeTags(tags, entry.Type, fp, entry.Tags)
	}

	b, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(KindyTagsCache, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	return err
}

func mergeTags(tagMap map[string]TagInfo, t local.KindyType, file string, tags []string) {
	for _, tag := range tags {
		lcTag := strings.ToLower(tag)
		tmp := tagMap[lcTag]
		tmp.Count++
		tmp.Entries = append(tmp.Entries, TagEntry{KindyType: t, KindyPath: filepath.ToSlash(file)})
		tagMap[lcTag] = tmp
	}
}

func kindyCacheFile(kind local.KindyType) string {
	switch kind {
	case local.KindyTypePost:
		return KindyPostsCache
	case local.KindyTypePhoto:
		return KindyPhotosCache
	case local.KindyTypeRepost:
		return KindyRepostsCache
	case local.KindyTypeLike:
		return KindyLikesCache
	case local.KindyTypeNote:
		return KindyNotesCache
	default:
		return ""
	}
}
