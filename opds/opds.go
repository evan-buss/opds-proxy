package opds

import (
	"encoding/xml"
	"io"
	"strings"
)

func ParseFeed(r io.Reader) (*Feed, error) {
	var feed Feed

	err := xml.NewDecoder(r).Decode(&feed)
	if err != nil {
		return nil, err
	}

	return &feed, nil
}

func (link Link) IsDownload() string {
	if link.Rel == "http://opds-spec.org/acquisition" {
		return link.Href
	}

	return ""
}

func (link Link) IsImage(category string) string {
	if strings.HasPrefix(link.TypeLink, "image") {
		if strings.Contains(link.Rel, category) {
			return link.Href
		}
	}

	return ""
}

func (link Link) IsNavigation() bool {
	return link.TypeLink == "application/atom+xml;type=feed;profile=opds-catalog"
}
