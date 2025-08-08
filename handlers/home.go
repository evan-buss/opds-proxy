package handlers

import (
	"io"
	"net/http"

	"github.com/evan-buss/opds-proxy/view"
)

type HomeLink struct {
	Title string
	URL   string
}

// Home returns a handler that renders the home page with provided links
func Home(links []HomeLink) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(links) == 1 {
			http.Redirect(w, r, "/feed?q="+links[0].URL, http.StatusFound)
			return
		}

		params := make([]view.HomeParams, len(links))
		for i, l := range links {
			params[i] = view.HomeParams{Title: l.Title, URL: l.URL}
		}

		view.Render(w, func(buf io.Writer) error { return view.Home(buf, params) })
	}
}
