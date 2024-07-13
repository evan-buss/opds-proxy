package html

import (
	"embed"
	"html/template"
	"io"

	"github.com/evan-buss/opds-proxy/opds"
	sprig "github.com/go-task/slim-sprig/v3"
)

//go:embed *
var files embed.FS

var (
	home  = parse("home.html")
	feed  = parse("feed.html", "partials/search.html")
	login = parse("login.html")
)

func parse(file ...string) *template.Template {
	file = append(file, "layout.html")
	return template.Must(
		template.New("layout.html").
			Funcs(sprig.FuncMap()).
			Funcs(template.FuncMap{
				"getKey": func(key string, d map[string]interface{}) interface{} {
					if val, ok := d[key]; ok {
						return val
					}
					return ""
				}},
			).
			ParseFS(files, file...),
	)
}

type LoginParams struct {
	ReturnURL string
}

func Login(w io.Writer, p LoginParams) error {
	return login.Execute(w, p)
}

type FeedParams struct {
	URL  string
	Feed *opds.Feed
}

func Feed(w io.Writer, p FeedParams) error {
	vm := convertFeed(&p)
	return feed.Execute(w, vm)
}

type FeedInfo struct {
	Title string
	URL   string
}

func Home(w io.Writer, vm []FeedInfo) error {
	return home.Execute(w, vm)
}

func StaticFiles() embed.FS {
	return files
}
