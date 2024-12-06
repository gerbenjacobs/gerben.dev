package internal

import (
	"encoding/json"
	"log/slog"
	"os"
	"sort"
	"strings"

	kindy "github.com/gerbenjacobs/gerben.dev"
)

type TagInfo struct {
	Count      int                          `json:"count"`
	Permalinks map[kindy.KindyType][]string `json:"permalinks"`
}

func GetTags() []string {
	tagMap, _ := getTagFile()
	var tags []string
	for tag, tagInfo := range tagMap {
		if tagInfo.Count > 1 {
			tags = append(tags, strings.ToLower(tag))
		}
	}
	sort.Strings(tags)
	return tags
}

func GetAllTags() map[string]TagInfo {
	tagMap, _ := getTagFile()
	return tagMap
}

func GetTag(tag string) map[string]TagInfo {
	tagMap := GetAllTags()
	tagInfo, ok := tagMap[tag]
	if !ok {
		return nil
	}
	return map[string]TagInfo{tag: tagInfo}
}

func getTagFile() (map[string]TagInfo, error) {
	var tagMap map[string]TagInfo
	b, err := os.ReadFile(kindy.KindyContentPath + "data/tags.json")
	if err != nil {
		slog.Warn("failed to read tags.json", "error", err)
		return nil, err
	}
	if err := json.Unmarshal(b, &tagMap); err != nil {
		slog.Warn("failed to unmarshal tags.json", "error", err)
		return nil, err
	}
	return tagMap, nil
}
