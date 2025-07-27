package opds

import "strings"

// LinkCategory represents different categories of links for filtering
type LinkCategory string

const (
	LinkCategoryAny       LinkCategory = ""
	LinkCategoryThumbnail LinkCategory = "thumbnail"
)

// Link represents a link to different resources
type Link struct {
	Rel                 string                `xml:"rel,attr"`
	Href                string                `xml:"href,attr"`
	TypeLink            string                `xml:"type,attr"`
	Title               string                `xml:"title,attr"`
	FacetGroup          string                `xml:"facetGroup,attr"`
	Count               int                   `xml:"count,attr"`
	Price               Price                 `xml:"price"`
	IndirectAcquisition []IndirectAcquisition `xml:"indirectAcquisition"`
}

// Price represents the book price
type Price struct {
	CurrencyCode string  `xml:"currencycode,attr"`
	Value        float64 `xml:",cdata"`
}

// IndirectAcquisition represents the link mostly for buying or borrowing a book
type IndirectAcquisition struct {
	TypeAcquisition     string                `xml:"type,attr"`
	IndirectAcquisition []IndirectAcquisition `xml:"indirectAcquisition"`
}

// Links represents a collection of Link objects with fluent filtering
type Links []Link

// Link methods

// IsDownload checks if the link is an acquisition/download link
func (l Link) IsDownload() bool {
	return l.Rel == AcquisitionFeedRel
}

// IsImage checks if the link is an image with optional category filtering
func (l Link) IsImage(category LinkCategory) bool {
	if strings.HasPrefix(l.TypeLink, "image") {
		return strings.Contains(l.Rel, string(category))
	}
	return false
}

// IsDataImage checks if the link is a data URI image
func (l Link) IsDataImage() bool {
	return strings.HasPrefix(l.Href, "data:")
}

// IsNavigation checks if the link is a navigation link
func (l Link) IsNavigation() bool {
	return l.TypeLink == NavigationFeedType || l.Rel == "subsection"
}

// IsThumbnail checks if the link is specifically a thumbnail image
func (l Link) IsThumbnail() bool {
	return l.IsImage(LinkCategoryThumbnail)
}

// HasRel checks if the link has the specified rel attribute
func (l Link) HasRel(rel string) bool {
	return l.Rel == rel
}

// HasType checks if the link type starts with the specified prefix
func (l Link) HasType(typePrefix string) bool {
	return strings.HasPrefix(l.TypeLink, typePrefix)
}

// Links methods

// Where filters links using the provided predicate function
func (links Links) Where(predicate func(Link) bool) Links {
	var filtered Links
	for _, link := range links {
		if predicate(link) {
			filtered = append(filtered, link)
		}
	}
	return filtered
}

// Downloads returns only download/acquisition links
func (links Links) Downloads() Links {
	return links.Where(func(link Link) bool {
		return link.IsDownload()
	})
}

// Images returns only image links, optionally filtered by category
func (links Links) Images(category ...LinkCategory) Links {
	cat := LinkCategoryAny
	if len(category) > 0 {
		cat = category[0]
	}
	return links.Where(func(link Link) bool {
		return link.IsImage(cat)
	})
}

// Navigation returns only navigation links
func (links Links) Navigation() Links {
	return links.Where(func(link Link) bool {
		return link.IsNavigation()
	})
}

// DataImages returns only data URI image links
func (links Links) DataImages() Links {
	return links.Where(func(link Link) bool {
		return link.IsDataImage()
	})
}

// First returns the first link that matches, or nil if none found
func (links Links) First() *Link {
	if len(links) > 0 {
		return &links[0]
	}
	return nil
}
