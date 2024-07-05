package html

import (
	"embed"
	"html/template"
	"io"
	"strings"

	"github.com/evan-buss/kobo-opds-proxy/opds"
)

//go:embed *
var files embed.FS

var (
	feed = parse("feed.html")
)

func parse(file string) *template.Template {
	return template.Must(template.New("layout.html").ParseFS(files, "layout.html", file))
}

type FeedParams struct {
	Feed *opds.Feed
}

type FeedViewModel struct {
	Title      string
	Search     string
	Navigation []NavigationViewModel
	Links      []LinkViewModel
}
type NavigationViewModel struct {
	Href  string
	Label string
}

type LinkViewModel struct {
	Title      string
	ImageURL   string
	Content    string
	Href       string
	IsDownload bool
}

func convertFeed(feed *opds.Feed) FeedViewModel {
	vm := FeedViewModel{
		Title:      feed.Title,
		Search:     "",
		Links:      make([]LinkViewModel, 0),
		Navigation: make([]NavigationViewModel, 0),
	}

	for _, link := range feed.Links {
		if link.Rel == "search" {
			vm.Search = link.Href
		}

		if link.TypeLink == "application/atom+xml;type=feed;profile=opds-catalog" {
			vm.Navigation = append(vm.Navigation, NavigationViewModel{
				Href:  link.Href,
				Label: strings.ToUpper(link.Rel[:1]) + link.Rel[1:],
			})
		}
	}

	for _, entry := range feed.Entries {
		vm.Links = append(vm.Links, constructLink(entry))
	}

	return vm
}

func constructLink(entry opds.Entry) LinkViewModel {
	vm := LinkViewModel{
		Title:   entry.Title,
		Content: entry.Content.Content,
	}

	for _, link := range entry.Links {
		if link.IsNavigation() || link.IsDownload() {
			vm.Href = link.Href
		}

		// Prefer the first "thumbnail" image we find
		if vm.ImageURL == "" && link.IsImage("thumbnail") {
			vm.ImageURL = link.Href
		}
	}

	// If we didn't find a thumbnail, use the first image we find
	if vm.ImageURL == "" {
		for _, link := range entry.Links {
			if link.IsImage("") {
				vm.ImageURL = link.Href
				break
			}
		}
	}

	return vm
}

func Feed(w io.Writer, p FeedParams, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}

	vm := convertFeed(p.Feed)

	return feed.ExecuteTemplate(w, partial, vm)
}

func StaticFiles() embed.FS {
	return files
}
