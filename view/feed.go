package view

import (
	"fmt"
	"html/template"
	"net/url"
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
func extractNavigationData(feed *opds.Feed, baseURL string) (NavigationData, error) {
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
		search, err := resolveHref(baseURL, searchLink.Href)
		if err != nil {
			return NavigationData{}, fmt.Errorf("failed to resolve search link: %w", err)
		}
		nav.Search = search
	}

	// Extract navigation links
	for _, link := range links.Navigation() {
		if link.Rel == "self" || // self link
			strings.Contains(link.Rel, "http") { // opds sort links
			continue // skip to save screen space
		}

		href, err := resolveHref(baseURL, link.Href)
		if err != nil {
			return NavigationData{}, fmt.Errorf("failed to resolve navigation link: %w", err)
		}
		nav.Navigation = append(nav.Navigation, NavigationViewModel{
			Href:  href,
			Label: strings.ToUpper(link.Rel[:1]) + link.Rel[1:],
		})
	}

	return nav, nil
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

func convertFeed(p *FeedParams) (FeedViewModel, error) {
	// Extract navigation data using shared function
	navData, err := extractNavigationData(p.Feed, p.URL)
	if err != nil {
		return FeedViewModel{}, fmt.Errorf("failed to extract navigation data: %w", err)
	}

	vm := FeedViewModel{
		Title:      p.Feed.Title,
		Search:     navData.Search,
		Navigation: navData.Navigation,
		Links:      make([]LinkViewModel, 0),
	}

	for _, entry := range p.Feed.Entries {
		link, err := constructLink(p.URL, entry)
		if err != nil {
			return FeedViewModel{}, fmt.Errorf("failed to construct link for entry %s: %w", entry.ID, err)
		}
		vm.Links = append(vm.Links, link)
	}

	return vm, nil
}

func constructLink(baseUrl string, entry opds.Entry) (LinkViewModel, error) {
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
		href, err := resolveHref(baseUrl, navLinks[0].Href)
		if err != nil {
			return LinkViewModel{}, fmt.Errorf("failed to resolve navigation link: %w", err)
		}
		vm.Href = url.QueryEscape(href)
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
			imageURL, err := resolveHref(baseUrl, imageLink.Href)
			if err != nil {
				return LinkViewModel{}, fmt.Errorf("failed to resolve image link: %w", err)
			}
			vm.ImageURL = imageURL
		}
	}

	return vm, nil
}

func resolveHref(feedUrl string, relativePath string) (string, error) {
	baseUrl, err := url.Parse(feedUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse feed URL %q: %w", feedUrl, err)
	}
	relativeUrl, err := url.Parse(relativePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse relative path %q: %w", relativePath, err)
	}

	return baseUrl.ResolveReference(relativeUrl).String(), nil
}
