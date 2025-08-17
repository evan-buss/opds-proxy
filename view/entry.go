package view

import (
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/evan-buss/opds-proxy/internal/formats"
)

const maxSummaryLength = 500

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// truncateSummary converts HTML to plain text and truncates to maxLength
func truncateSummary(htmlText string, maxLength int) string {
	// Strip HTML tags and unescape entities to get plain text
	plainText := htmlTagRegex.ReplaceAllString(htmlText, "")
	plainText = html.UnescapeString(plainText)
	plainText = strings.TrimSpace(plainText)

	if utf8.RuneCountInString(plainText) <= maxLength {
		return plainText
	}

	runes := []rune(plainText)
	truncated := runes[:maxLength]

	// Find the last space before maxLength to avoid cutting words
	lastSpace := len(truncated) - 1
	for i := len(truncated) - 1; i >= 0; i-- {
		if truncated[i] == ' ' {
			lastSpace = i
			break
		}
	}

	// If we found a space within reasonable distance, use it
	if lastSpace > maxLength-50 {
		return string(truncated[:lastSpace]) + "..."
	}

	return string(truncated) + "..."
}

// EntryViewModel is the data passed to the entry.html template.
type EntryViewModel struct {
	Title           string
	Author          string
	Content         template.HTML
	DownloadLinks   []EntryLinkViewModel
	NavigationLinks []EntryLinkViewModel
	FeedURL         string
	ImageURL        string
	ImageData       template.URL
	Search          string
	Navigation      []NavigationViewModel
}

// EntryLinkViewModel is a single link in the entry.html template.
type EntryLinkViewModel struct {
	Title    string
	Href     string
	TypeLink string
	Subtext  string
}

func constructEntryVM(params EntryParams) (EntryViewModel, error) {
	// Extract navigation data using shared function from feed.go
	navData, err := extractNavigationData(params.Feed, params.URL)
	if err != nil {
		return EntryViewModel{}, fmt.Errorf("failed to extract navigation data: %w", err)
	}

	vm := EntryViewModel{
		Title:           params.Entry.Title,
		DownloadLinks:   []EntryLinkViewModel{},
		NavigationLinks: []EntryLinkViewModel{},
		FeedURL:         params.URL,
		Content:         template.HTML(truncateSummary(params.Entry.SummaryText(), maxSummaryLength)),
		Author:          strings.Join(params.Entry.AuthorNames(), " & "),
		Search:          navData.Search,
		Navigation:      navData.Navigation,
		// ImageURL: resolveHref(params.URL, params.Entry.Image()),
	}

	imageLink := params.Entry.Image()
	if imageLink != nil {
		if imageLink.IsDataImage() {
			vm.ImageData = template.URL(imageLink.Href)
		} else {
			imageURL, err := resolveHref(params.URL, imageLink.Href)
			if err != nil {
				return EntryViewModel{}, fmt.Errorf("failed to resolve image link: %w", err)
			}
			vm.ImageURL = imageURL
		}
	}

	links := params.Entry.GetLinks()

	for _, link := range links.Navigation() {
		href, err := resolveHref(params.URL, link.Href)
		if err != nil {
			return EntryViewModel{}, fmt.Errorf("failed to resolve navigation link: %w", err)
		}
		vm.NavigationLinks = append(vm.NavigationLinks, EntryLinkViewModel{
			Title:    link.Title,
			Href:     href,
			TypeLink: link.TypeLink,
		})
	}

	for _, link := range links.Downloads() {
		// Use link's title if present, otherwise generate custom title
		var title string
		if link.Title != "" {
			title = link.Title
		} else {
			title = formats.GetMimeTypeLabel(link.TypeLink) + " Format"
		}

		if link.TypeLink == params.DeviceType.GetPreferredFormat().MimeType {
			title += " (Recommended)"
		}

		format, exists := formats.FormatByMimeType(link.TypeLink)
		if !exists {
			href, err := resolveHref(params.URL, link.Href)
			if err != nil {
				return EntryViewModel{}, fmt.Errorf("failed to resolve download link: %w", err)
			}
			vm.DownloadLinks = append(vm.DownloadLinks, EntryLinkViewModel{
				Title:    title,
				Href:     href,
				TypeLink: link.TypeLink,
			})
			continue
		}
		converter := params.ConverterManager.GetConverterForDevice(params.DeviceType, format)
		subtext := ""
		if converter != nil {
			// If the converter handles this format, we can add a note
			subtext += "Automatically converted to " + params.DeviceType.GetPreferredFormat().Label + ". "
		}

		href, err := resolveHref(params.URL, link.Href)
		if err != nil {
			return EntryViewModel{}, fmt.Errorf("failed to resolve download link: %w", err)
		}
		vm.DownloadLinks = append(vm.DownloadLinks, EntryLinkViewModel{
			Title:    title,
			Href:     href,
			TypeLink: link.TypeLink,
			Subtext:  subtext,
		})
	}

	return vm, nil
}
