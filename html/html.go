package html

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/evan-buss/opds-proxy/convert"
	"github.com/evan-buss/opds-proxy/internal/device"
	"github.com/evan-buss/opds-proxy/opds"
	sprig "github.com/go-task/slim-sprig/v3"
)

//go:embed *
var files embed.FS

var (
	home  = parse("home.html")
	login = parse("login.html")
	feed  = parse("feed.html", "partials/search.html")
	entry = parse("entry.html", "partials/search.html")
)

func parse(file ...string) *template.Template {
	file = append(file, "layout.html")
	return template.Must(
		template.New("layout.html").
			Funcs(sprig.FuncMap()).
			Funcs(template.FuncMap{
				"getKey": func(key string, d map[string]any) any {
					if val, ok := d[key]; ok {
						return val
					}
					return ""
				}},
			).
			ParseFS(files, file...),
	)
}

type HomeParams struct {
	Title string
	URL   string
}

func Home(w io.Writer, vm []HomeParams) error {
	return home.Execute(w, vm)
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

type EntryParams struct {
	URL              string
	Feed             *opds.Feed
	Entry            opds.Entry
	DeviceType       device.DeviceType
	ConverterManager *convert.ConverterManager
}

func Entry(w io.Writer, p EntryParams) error {
	fmt.Println("Converting entry:", p.Entry.Title)
	vm := constructEntryVM(p)
	return entry.Execute(w, vm)
}

func StaticFiles() embed.FS {
	return files
}
