package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/evan-buss/kobo-opds-proxy/html"
	"github.com/evan-buss/kobo-opds-proxy/opds"
)

const baseUrl = "http://calibre.terminus"

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /", handleFeed())
	router.HandleFunc("GET /acquisition/{path...}", handleAcquisition())
	router.Handle("GET /static/", http.FileServer(http.FS(html.StaticFiles())))

	slog.Info("Starting server", slog.String("port", "8080"))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opdsUrl := baseUrl + r.URL.RequestURI()

		slog.Info("Fetching feed", slog.String("path", opdsUrl))

		feed, err := fetchFeed(opdsUrl)
		if err != nil {
			slog.Error("Failed to parse feed", slog.String("path", r.URL.RawPath), slog.Any("error", err))
			http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
			return
		}

		err = html.Feed(w, html.FeedParams{Feed: feed}, "")
		if err != nil {
			slog.Error("Failed to render feed", slog.String("path", r.URL.RawPath), slog.Any("error", err))
			http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
			return
		}
	}
}

func handleAcquisition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opdsUrl := baseUrl + "/" + r.PathValue("path")
		resp, err := fetchFromUrl(opdsUrl)
		if err != nil {
			slog.Error("Failed to fetch acquisition", slog.String("path", opdsUrl), slog.Any("error", err))
			http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
			return
		}

		for k, v := range resp.Header {
			w.Header()[k] = v
		}

		io.Copy(w, resp.Body)
	}
}

func fetchFromUrl(url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("user", "password")

	return client.Do(req)
}

func fetchFeed(url string) (*opds.Feed, error) {
	r, err := fetchFromUrl(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return opds.ParseFeed(r.Body)
}
