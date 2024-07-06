package html

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/evan-buss/kobo-opds-proxy/opds"
)

//go:embed *
var files embed.FS

var (
	home = parse("home.html")
	feed = parse("feed.html")
)

func parse(file string) *template.Template {
	return template.Must(template.New("layout.html").ParseFS(files, "layout.html", file))
}

type FeedParams struct {
	URL  string
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
	Author     string
	ImageURL   string
	Content    string
	Href       string
	IsDownload bool
}

func convertFeed(url string, feed *opds.Feed) FeedViewModel {
	vm := FeedViewModel{
		Title:      feed.Title,
		Search:     "",
		Links:      make([]LinkViewModel, 0),
		Navigation: make([]NavigationViewModel, 0),
	}

	for _, link := range feed.Links {
		if link.Rel == "search" {
			vm.Search = resolveHref(url, link.Href)
		}

		if link.TypeLink == "application/atom+xml;type=feed;profile=opds-catalog" {
			vm.Navigation = append(vm.Navigation, NavigationViewModel{
				Href:  resolveHref(url, link.Href),
				Label: strings.ToUpper(link.Rel[:1]) + link.Rel[1:],
			})
		}
	}

	for _, entry := range feed.Entries {
		vm.Links = append(vm.Links, constructLink(url, entry))
	}

	return vm
}

func constructLink(url string, entry opds.Entry) LinkViewModel {
	vm := LinkViewModel{
		Title:   entry.Title,
		Content: entry.Content.Content,
	}

	authors := make([]string, 0)
	for _, author := range entry.Author {
		authors = append(authors, author.Name)
	}
	vm.Author = strings.Join(authors, " & ")

	for _, link := range entry.Links {
		vm.IsDownload = link.IsDownload()
		if link.IsNavigation() || link.IsDownload() {
			vm.Href = resolveHref(url, link.Href)
		}

		// Prefer the first "thumbnail" image we find
		if vm.ImageURL == "" && link.IsImage("thumbnail") {
			vm.ImageURL = resolveHref(url, link.Href)
		}
	}

	// If we didn't find a thumbnail, use the first image we find
	if vm.ImageURL == "" {
		for _, link := range entry.Links {
			if link.IsImage("") {
				vm.ImageURL = resolveHref(url, link.Href)
				break
			}
		}
	}

	return vm
}
func resolveHref(feedUrl string, relativePath string) string {
	baseUrl, err := url.Parse(feedUrl)
	if err != nil {
		log.Fatal(err)
	}
	relativeUrl, err := url.Parse(relativePath)
	if err != nil {
		log.Fatal(err)
	}

	resolved := baseUrl.ResolveReference(relativeUrl).String()
	fmt.Println("Resolved URL: ", resolved)
	return resolved
}

func Feed(w io.Writer, p FeedParams, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}

	vm := convertFeed(p.URL, p.Feed)

	return feed.ExecuteTemplate(w, partial, vm)
}

type FeedInfo struct {
	Title string
	URL   string
}

func Home(w io.Writer, p []FeedInfo) error {
	return home.ExecuteTemplate(w, "layout.html", p)
}

func StaticFiles() embed.FS {
	return files
}
