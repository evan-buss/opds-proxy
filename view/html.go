package view

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"net/http"

	"github.com/evan-buss/opds-proxy/convert"
	"github.com/evan-buss/opds-proxy/internal/device"
	"github.com/evan-buss/opds-proxy/opds"
	sprig "github.com/go-task/slim-sprig/v3"
)

//go:embed *.html partials/*.html static/*
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

// Render safely writes HTML to the ResponseWriter.
// It first renders the template/content into a buffer so that:
// 1) We avoid sending partial responses if rendering fails midway.
// 2) We can choose the correct HTTP status code on errors before any bytes are written.
// Only after a successful render do we set the Content-Type and write the body.
func Render(w http.ResponseWriter, render func(io.Writer) error) {
	var buf bytes.Buffer
	if err := render(&buf); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
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
	vm, err := convertFeed(&p)
	if err != nil {
		return err
	}
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
	vm, err := constructEntryVM(p)
	if err != nil {
		return err
	}
	return entry.Execute(w, vm)
}

func StaticFiles() embed.FS {
	return files
}
