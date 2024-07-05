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

func (link Link) IsDownload() bool {
	return link.Rel == "http://opds-spec.org/acquisition"
}

func (link Link) IsImage(category string) bool {
	if strings.HasPrefix(link.TypeLink, "image") {
		return strings.Contains(link.Rel, category)
	}

	return false
}

func (link Link) IsNavigation() bool {
	return link.TypeLink == "application/atom+xml;type=feed;profile=opds-catalog"
}
