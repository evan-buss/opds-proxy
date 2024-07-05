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
	feed = parse("feed.html")
)

func parse(file string) *template.Template {
	return template.Must(template.New("layout.html").ParseFS(files, "layout.html", file))
}

type FeedParams struct {
	Feed *opds.Feed
}

func Feed(w io.Writer, p FeedParams, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}

	return feed.ExecuteTemplate(w, partial, p)
}

func StaticFiles() embed.FS {
	return files
}
