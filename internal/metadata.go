package internal

import (
	"html"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
)

var (
	TitlifyLength     = 51
	DescriptifyLength = 200
)

// Metadata can be created for every page
// it holds information for title, description and opengraph data
type Metadata struct {
	Env         string
	Title       string
	Description string
	Permalink   string
	Image       string
	Kindy       *local.Kindy
	SourceLink  string
}

func Titlify(title string) string {
	title = html.UnescapeString(title)
	title = strings.Join(strings.Fields(title), " ")
	if len(title) > TitlifyLength {
		return title[:TitlifyLength] + "..."
	}
	return title
}

func Descriptify(description string) string {
	// removes all redundant spaces
	description = strings.Join(strings.Fields(description), " ")

	if len(description) > DescriptifyLength {
		return description[:DescriptifyLength] + "..."
	}
	return description
}
