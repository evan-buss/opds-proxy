package html 

import (
	"html/template"
	"io"
	"github.com/opds-community/libopds2-go/opds1"
)

var (
	feed = parse("feed.html")
)

func parse(file string) *template.Template {
	return template.Must(template.New("layout.html").ParseFiles("html/layout.html", "html/" + file))
}

type FeedParams struct {
	Feed *opds1.Feed
}

func Feed(w io.Writer, p FeedParams, partial string) error {
	if (partial == "" ) {
		partial = "layout.html"
	}

	return feed.ExecuteTemplate(w, partial, p)
}