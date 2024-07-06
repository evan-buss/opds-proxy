package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/evan-buss/kobo-opds-proxy/html"
	"github.com/evan-buss/kobo-opds-proxy/opds"
)

func main() {
	server := NewServer()

	slog.Info("Starting server", slog.String("port", server.addr))
	log.Fatal(http.ListenAndServe(server.addr, server.router))
}

func handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Rendering home page")
		feeds := []html.FeedInfo{
			{
				Title: "Evan's Library",
				URL:   "http://calibre.terminus/opds",
			},
			{
				Title: "Project Gutenberg",
				URL:   "https://m.gutenberg.org/ebooks.opds/",
			},
		}
		html.Home(w, feeds, partial(r))
	}
}

func handleFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryURL := r.URL.Query().Get("q")
		if queryURL == "" {
			http.Error(w, "No feed specified", http.StatusBadRequest)
			return
		}

		parsedUrl, err := url.PathUnescape(queryURL)
		if err != nil {
			handleError(r, w, "Failed to parse URL", err)
			return
		}
		queryURL = parsedUrl

		searchTerm := r.URL.Query().Get("search")
		if searchTerm != "" {
			fmt.Println("Search term", searchTerm)
			queryURL = replaceSearchPlaceHolder(queryURL, searchTerm)
		}

		resp, err := fetchFromUrl(queryURL)
		if err != nil {
			handleError(r, w, "Failed to fetch feed", err)
			return
		}

		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		mimeType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			handleError(r, w, "Failed to parse content type", err)
		}

		if mimeType == "application/atom+xml" {
			feed, err := opds.ParseFeed(resp.Body)
			if err != nil {
				handleError(r, w, "Failed to parse feed", err)
				return
			}

			feedParams := html.FeedParams{
				URL:  queryURL,
				Feed: feed,
			}

			err = html.Feed(w, feedParams, partial(r))
			if err != nil {
				handleError(r, w, "Failed to render feed", err)
				return
			}
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

func handleError(r *http.Request, w http.ResponseWriter, message string, err error) {
	slog.Error(message, slog.String("path", r.URL.RawPath), slog.Any("error", err))
	http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
}

func replaceSearchPlaceHolder(url string, searchTerm string) string {
	return strings.Replace(url, "{searchTerms}", searchTerm, 1)
}

func partial(req *http.Request) string {
	return req.URL.Query().Get("partial")
}
