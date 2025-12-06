package internal

import (
	"encoding/xml"

	local "github.com/gerbenjacobs/gerben.dev"
)

var sitemapPages = []*xmlSitemapItem{
	{Loc: "https://gerben.dev/", ChangeFreq: "daily", Priority: "1.0"},
	{Loc: "https://gerben.dev/projects", ChangeFreq: "monthly", Priority: "0.5"},
	{Loc: "https://gerben.dev/posts/", ChangeFreq: "weekly", Priority: "0.8"},
	{Loc: "https://gerben.dev/photos/", ChangeFreq: "weekly", Priority: "0.8"},
	{Loc: "https://gerben.dev/photos/featured", ChangeFreq: "monthly", Priority: "0.7"},
	{Loc: "https://gerben.dev/timeline", ChangeFreq: "daily", Priority: "1.0"},
	{Loc: "https://gerben.dev/previously", ChangeFreq: "daily", Priority: "0.6"},
	{Loc: "https://gerben.dev/changelog", ChangeFreq: "weekly", Priority: "0.5"},
	{Loc: "https://gerben.dev/sitemap", ChangeFreq: "daily", Priority: "0.6"},
	{Loc: "https://gerben.dev/listening", ChangeFreq: "daily", Priority: "0.6"},
	{Loc: "https://gerben.dev/collection", ChangeFreq: "monthly", Priority: "0.4"},
	{Loc: "https://gerben.dev/poems", ChangeFreq: "yearly", Priority: "0.4"},
}

type xmlSitemap struct {
	XMLName xml.Name          `xml:"urlset"`
	URL     []*xmlSitemapItem `xml:"url"`
	Xmlns   string            `xml:"xmlns,attr"`
}

type xmlSitemapItem struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func CreateSitemapXML() ([]byte, error) {
	var entries []local.Kindy
	for _, kindyType := range KindyTypes {
		ent, err := GetKindyCacheByType(kindyType)
		if err != nil {
			return nil, err
		}
		entries = append(entries, ent...)
	}

	// start with sitemapPages
	var items = sitemapPages

	// add all Kindy content
	for _, entry := range entries {
		url := "https://gerben.dev" + entry.Permalink
		item := &xmlSitemapItem{
			Loc:        url,
			LastMod:    entry.PublishedAt.Format("2006-01-02"),
			ChangeFreq: "yearly",
			Priority:   "0.5",
		}
		items = append(items, item)
	}

	// add Tags
	for _, tag := range GetTags() {
		url := "https://gerben.dev/tags/" + tag
		item := &xmlSitemapItem{
			Loc:        url,
			ChangeFreq: "weekly",
			Priority:   "0.4",
		}
		items = append(items, item)
	}

	sitemap := &xmlSitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URL:   items,
	}

	return xml.Marshal(sitemap)
}
