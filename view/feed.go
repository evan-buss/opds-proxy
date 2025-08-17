package view

import (
	"html/template"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/evan-buss/opds-proxy/opds"
)

type FeedViewModel struct {
	Title      string
	Search     string
	Navigation []NavigationViewModel
	Links      []LinkViewModel
}

// NavigationData contains the common navigation and search data
type NavigationData struct {
	Search     string
	Navigation []NavigationViewModel
}

// extractNavigationData extracts navigation and search links from a feed
func extractNavigationData(feed *opds.Feed, baseURL string) NavigationData {
	nav := NavigationData{
		Search:     "",
		Navigation: make([]NavigationViewModel, 0),
	}

	links := feed.GetLinks()

	// Find search link
	searchLinks := links.Where(func(link opds.Link) bool {
		return link.Rel == "search"
	})
	if searchLink := searchLinks.First(); searchLink != nil {
		nav.Search = resolveHref(baseURL, searchLink.Href)
	}

	// Extract navigation links
	for _, link := range links.Navigation() {
		if link.Rel == "self" || // self link
			strings.Contains(link.Rel, "http") { // opds sort links
			continue // skip to save screen space
		}

		nav.Navigation = append(nav.Navigation, NavigationViewModel{
			Href:  resolveHref(baseURL, link.Href),
			Label: strings.ToUpper(link.Rel[:1]) + link.Rel[1:],
		})
	}

	return nav
}

type NavigationViewModel struct {
	Href  string
	Label string
}

type LinkViewModel struct {
	Title     string
	Author    string
	ImageURL  string
	ImageData template.URL // Base64 encoded image data
	Content   string
	Href      string
	EntryID   string
}

func convertFeed(p *FeedParams) FeedViewModel {
	// Extract navigation data using shared function
	navData := extractNavigationData(p.Feed, p.URL)

	vm := FeedViewModel{
		Title:      p.Feed.Title,
		Search:     navData.Search,
		Navigation: navData.Navigation,
		Links:      make([]LinkViewModel, 0),
	}

	for _, entry := range p.Feed.Entries {
		vm.Links = append(vm.Links, constructLink(p.URL, entry))
	}

	return vm
}

func constructLink(baseUrl string, entry opds.Entry) LinkViewModel {
	vm := LinkViewModel{
		Title:   entry.Title,
		Content: entry.Content.Content,
		// Href:    "/entry?q=" + url.QueryEscape(baseUrl) + "&id=" + entry.ID,
	}

	authors := make([]string, 0)
	for _, author := range entry.Author {
		authors = append(authors, author.Name)
	}
	vm.Author = strings.Join(authors, " & ")

	navLinks := entry.GetLinks().Navigation()

	// If there is 1 link and it's a navigation link, don't link to the entry details page
	if len(navLinks) == 1 {
		vm.Href = url.QueryEscape(resolveHref(baseUrl, navLinks[0].Href))
		vm.EntryID = ""
	} else {
		// Otherwise, link to the entry details page
		vm.Href = url.QueryEscape(baseUrl)
		vm.EntryID = entry.ID
	}

	imageLink := entry.Thumbnail()
	if imageLink != nil {
		if imageLink.IsDataImage() {
			vm.ImageData = template.URL(imageLink.Href)
		} else {
			vm.ImageURL = resolveHref(baseUrl, imageLink.Href)
		}
	}

	return vm
}

func resolveHref(feedUrl string, relativePath string) string {
	baseUrl, err := url.Parse(feedUrl)
	if err != nil {
		slog.Error("failed to parse feed URL", slog.Any("error", err), slog.String("url", feedUrl))
		os.Exit(1)
	}
	relativeUrl, err := url.Parse(relativePath)
	if err != nil {
		slog.Error("failed to parse relative path", slog.Any("error", err), slog.String("path", relativePath))
		os.Exit(1)
	}

	return baseUrl.ResolveReference(relativeUrl).String()
}
