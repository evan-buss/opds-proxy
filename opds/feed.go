package opds

// OPDS navigation feed type constant
const NavigationFeedType string = "profile=opds-catalog"

// OPDS acquisition constant
const AcquisitionFeedRel string = "http://opds-spec.org/acquisition"

// Feed represents the root element for acquisition or navigation feeds
type Feed struct {
	ID           string  `xml:"id"`
	Title        string  `xml:"title"`
	Updated      Time    `xml:"updated"`
	Entries      []Entry `xml:"entry"`
	Links        []Link  `xml:"link"`
	TotalResults int     `xml:"totalResults"`
	ItemsPerPage int     `xml:"itemsPerPage"`
}

// GetLinks returns the links as a fluent Links type for filtering
func (f Feed) GetLinks() Links {
	return Links(f.Links)
}

// IsAcquisitionFeed checks if this feed contains entries with acquisition links
func (f *Feed) IsAcquisitionFeed() bool {
	for _, entry := range f.Entries {
		acquisitionFeeds := entry.GetLinks().Where(func(link Link) bool {
			return link.IsDownload()
		})

		if len(acquisitionFeeds) > 0 {
			return true
		}
	}

	return false
}

// IsNavigationFeed checks if this feed contains entries with navigation links
func (f *Feed) IsNavigationFeed() bool {
	for _, entry := range f.Entries {
		navigationLinks := entry.GetLinks().Where(func(link Link) bool {
			return link.IsNavigation()
		})

		if len(navigationLinks) > 0 {
			return true
		}
	}

	return false
}
