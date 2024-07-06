package html

import (
	"embed"
	"html/template"
	"io"

	"github.com/evan-buss/kobo-opds-proxy/opds"
)

//go:embed *
var files embed.FS

var (
	home = parse("home.html")
	feed = parse("feed.html", "partials/search.html")
)

func parse(file ...string) *template.Template {
	file = append(file, "layout.html")
	return template.Must(template.New("layout.html").ParseFS(files, file...))
}

type FeedParams struct {
	URL  string
	Feed *opds.Feed
}

func Feed(w io.Writer, p FeedParams, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}

	vm := convertFeed(&p)

	return feed.ExecuteTemplate(w, partial, vm)
}

type FeedInfo struct {
	Title string
	URL   string
}

func Home(w io.Writer, vm []FeedInfo, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}
	return home.ExecuteTemplate(w, partial, vm)
}

func StaticFiles() embed.FS {
	return files
}
