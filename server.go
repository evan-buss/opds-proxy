package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evan-buss/opds-proxy/convert"
	"github.com/evan-buss/opds-proxy/html"
	"github.com/evan-buss/opds-proxy/opds"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Server struct {
	addr   string
	router *http.ServeMux
}

func NewServer(config *config) *Server {
	router := http.NewServeMux()
	router.HandleFunc("GET /{$}", handleHome(config.Feeds))
	router.HandleFunc("GET /feed", handleFeed("tmp/"))
	router.Handle("GET /static/", http.FileServer(http.FS(html.StaticFiles())))

	return &Server{
		addr:   ":" + config.Port,
		router: router,
	}
}

func (s *Server) Serve() {
	slog.Info("Starting server", slog.String("port", s.addr))
	log.Fatal(http.ListenAndServe(s.addr, s.router))
}

func handleHome(feeds []feedConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vmFeeds := make([]html.FeedInfo, len(feeds))
		for i, feed := range feeds {
			vmFeeds[i] = html.FeedInfo{
				Title: feed.Name,
				URL:   feed.Url,
			}
		}

		html.Home(w, vmFeeds, partial(r))
	}
}

func handleFeed(dir string) http.HandlerFunc {
	kepubConverter := &convert.KepubConverter{}
	mobiConverter := &convert.MobiConverter{}

	return func(w http.ResponseWriter, r *http.Request) {
		queryURL := r.URL.Query().Get("q")
		if queryURL == "" {
			http.Error(w, "No feed specified", http.StatusBadRequest)
			return
		}

		parsedUrl, err := url.PathUnescape(queryURL)
		queryURL = parsedUrl
		if err != nil {
			handleError(r, w, "Failed to parse URL", err)
			return
		}

		searchTerm := r.URL.Query().Get("search")
		if searchTerm != "" {
			fmt.Println("Search term", searchTerm)
			queryURL = replaceSearchPlaceHolder(queryURL, searchTerm)
		}

		resp, err := fetchFromUrl(queryURL)
		if err != nil {
			handleError(r, w, "Failed to fetch", err)
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

		if mimeType != convert.EPUB_MIME {
			for k, v := range resp.Header {
				w.Header()[k] = v
			}

			io.Copy(w, resp.Body)
			return
		}

		if strings.Contains(r.Header.Get("User-Agent"), "Kobo") && kepubConverter.Available() {
			epubFile := filepath.Join(dir, parseFileName(resp))
			downloadFile(epubFile, resp)

			kepubFile := filepath.Join(dir, strings.Replace(parseFileName(resp), ".epub", ".kepub.epub", 1))
			kepubConverter.Convert(epubFile, kepubFile)
			if err != nil {
				handleError(r, w, "Failed to convert to kepub", err)
			}

			outFile, _ := os.Open(kepubFile)
			defer outFile.Close()

			outInfo, _ := outFile.Stat()

			w.Header().Set("Content-Length", fmt.Sprintf("%d", outInfo.Size()))
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filepath.Base(kepubFile)}))
			w.Header().Set("Content-Type", convert.EPUB_MIME)

			io.Copy(w, outFile)

			os.Remove(epubFile)
			os.Remove(kepubFile)

			return
		}

		if strings.Contains(r.Header.Get("User-Agent"), "Kindle") && mobiConverter.Available() {
			epubFile := filepath.Join(dir, parseFileName(resp))
			downloadFile(epubFile, resp)

			mobiFile := filepath.Join(dir, strings.Replace(parseFileName(resp), ".epub", ".mobi", 1))
			err := mobiConverter.Convert(epubFile, mobiFile)
			if err != nil {
				handleError(r, w, "Failed to convert to mobi", err)
				return
			}

			outFile, _ := os.Open(mobiFile)
			defer outFile.Close()

			outInfo, _ := outFile.Stat()

			w.Header().Set("Content-Length", fmt.Sprintf("%d", outInfo.Size()))
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filepath.Base(mobiFile)}))
			w.Header().Set("Content-Type", convert.MOBI_MIME)

			io.Copy(w, outFile)

			os.Remove(epubFile)
			os.Remove(mobiFile)

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
	req.SetBasicAuth("public", "evanbuss")

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

func downloadFile(path string, resp *http.Response) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func parseFileName(resp *http.Response) string {
	contentDisposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		log.Fatal(err)
	}

	return params["filename"]
}
