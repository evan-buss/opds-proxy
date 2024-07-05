package opds

import (
	"time"
)

// Feed root element for acquisition or navigation feed
type Feed struct {
	ID           string    `xml:"id"`
	Title        string    `xml:"title"`
	Updated      time.Time `xml:"updated"`
	Entries      []Entry   `xml:"entry"`
	Links        []Link    `xml:"link"`
	TotalResults int       `xml:"totalResults"`
	ItemsPerPage int       `xml:"itemsPerPage"`
}

// Link link to different resources
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

// Author represent the feed author or the entry author
type Author struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

// Entry an atom entry in the feed
type Entry struct {
	Title      string     `xml:"title"`
	ID         string     `xml:"id"`
	Identifier string     `xml:"identifier"`
	Updated    *time.Time `xml:"updated"`
	Rights     string     `xml:"rights"`
	Publisher  string     `xml:"publisher"`
	Author     []Author   `xml:"author,omitempty"`
	Language   string     `xml:"language"`
	Issued     string     `xml:"issued"` // Check for format
	Published  *time.Time `xml:"published"`
	Category   []Category `xml:"category,omitempty"`
	Links      []Link     `xml:"link,omitempty"`
	Summary    Content    `xml:"summary"`
	Content    Content    `xml:"content"`
	Series     []Serie    `xml:"Series"`
}

// Content content tag in an entry, the type will be html or text
type Content struct {
	Content     string `xml:",cdata"`
	ContentType string `xml:"type,attr"`
}

// Category represent the book category with scheme and term to machine
// handling
type Category struct {
	Scheme string `xml:"scheme,attr"`
	Term   string `xml:"term,attr"`
	Label  string `xml:"label,attr"`
}

// Price represent the book price
type Price struct {
	CurrencyCode string  `xml:"currencycode,attr"`
	Value        float64 `xml:",cdata"`
}

// IndirectAcquisition represent the link mostly for buying or borrowing
// a book
type IndirectAcquisition struct {
	TypeAcquisition     string                `xml:"type,attr"`
	IndirectAcquisition []IndirectAcquisition `xml:"indirectAcquisition"`
}

// Serie store serie information from schema.org
type Serie struct {
	Name     string  `xml:"name,attr"`
	URL      string  `xml:"url,attr"`
	Position float32 `xml:"position,attr"`
}
