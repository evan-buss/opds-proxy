package opds

// Author represents the feed author or the entry author
type Author struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

// Entry represents an atom entry in the feed
type Entry struct {
	Title      string        `xml:"title"`
	ID         string        `xml:"id"`
	Identifier string        `xml:"identifier"`
	Updated    *Time `xml:"updated"`
	Rights     string        `xml:"rights"`
	Publisher  string        `xml:"publisher"`
	Author     []Author      `xml:"author,omitempty"`
	Language   string        `xml:"language"`
	Issued     string        `xml:"issued"` // Check for format
	Published  *Time `xml:"published"`
	Category   []Category    `xml:"category,omitempty"`
	Links      []Link        `xml:"link,omitempty"`
	Summary    Content       `xml:"summary"`
	Content    Content       `xml:"content"`
	Series     []Serie       `xml:"Series"`
}

// Content represents content tag in an entry, the type will be html or text
type Content struct {
	Content     string `xml:",innerxml"`
	ContentType string `xml:"type,attr"`
}

// Category represents the book category with scheme and term for machine handling
type Category struct {
	Scheme string `xml:"scheme,attr"`
	Term   string `xml:"term,attr"`
	Label  string `xml:"label,attr"`
}

// Serie stores serie information from schema.org
type Serie struct {
	Name     string  `xml:"name,attr"`
	URL      string  `xml:"url,attr"`
	Position float32 `xml:"position,attr"`
}

// Entry methods

// GetLinks returns the links as a fluent Links type for filtering
func (e Entry) GetLinks() Links {
	return Links(e.Links)
}

// AuthorNames returns a slice of author names for the entry
func (e Entry) AuthorNames() []string {
	authors := make([]string, 0, len(e.Author))
	for _, author := range e.Author {
		authors = append(authors, author.Name)
	}
	return authors
}

// Thumbnail returns the first thumbnail image link for the entry
func (e Entry) Thumbnail() *Link {
	return e.GetLinks().Images(LinkCategoryThumbnail).First()
}

// Image returns the first image link for the entry
func (e Entry) Image() *Link {
	return e.GetLinks().Images().First()
}

// SummaryText returns the text content from summary or content fields
func (e Entry) SummaryText() string {
	if e.Summary.Content != "" {
		return e.Summary.Content
	}
	if e.Content.Content != "" {
		return e.Content.Content
	}
	return ""
}