package handlers

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"log/slog"

	"github.com/evan-buss/opds-proxy/convert"
	"github.com/evan-buss/opds-proxy/internal/auth"
	"github.com/evan-buss/opds-proxy/internal/device"
	"github.com/evan-buss/opds-proxy/internal/formats"
	"github.com/evan-buss/opds-proxy/internal/httpx"
	"github.com/evan-buss/opds-proxy/internal/reqctx"
	"github.com/evan-buss/opds-proxy/opds"
	"github.com/evan-buss/opds-proxy/view"
	"github.com/gorilla/securecookie"
)

type FeedHandler struct {
	outputDir  string
	feeds      []auth.FeedConfig
	s          *securecookie.SecureCookie
	debug      bool
	converters *convert.ConverterManager
	mu         sync.Mutex
}

func Feed(outputDir string, feeds []auth.FeedConfig, s *securecookie.SecureCookie, debug bool) http.HandlerFunc {
	h := &FeedHandler{
		outputDir:  outputDir,
		feeds:      feeds,
		s:          s,
		debug:      debug,
		converters: convert.NewConverterManager(),
	}
	return h.ServeHTTP
}

func (h *FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queryURL := r.URL.Query().Get("q")
	downloadType := r.URL.Query().Get("t")
	if queryURL == "" {
		http.Error(w, "No feed specified", http.StatusBadRequest)
		return
	}

	resolvedURL, err := h.resolveQueryURL(queryURL, r.URL.Query().Get("search"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse URL %q: %v", queryURL, err), http.StatusBadRequest)
		return
	}

	var customUserAgent string
	if parsedReqURL, err := url.Parse(resolvedURL); err == nil {
		for _, feed := range h.feeds {
			if parsedFeedURL, err := url.Parse(feed.Url); err == nil {
				if parsedFeedURL.Hostname() == parsedReqURL.Hostname() && feed.UserAgent != "" {
					customUserAgent = feed.UserAgent
					break
				}
			}
		}
	}

	creds := auth.GetCredentials(resolvedURL, r, h.feeds, h.s)
	resp, err := httpx.Fetch(resolvedURL, 10, func(req *http.Request) {
		if creds != nil {
			req.SetBasicAuth(creds.Username, creds.Password)
		}
		if customUserAgent != "" {
			req.Header.Set("User-Agent", customUserAgent)
		}
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch %q: %v", resolvedURL, err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		http.Redirect(w, r, "/auth?return="+r.URL.String(), http.StatusFound)
		return
	}

	mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if mimeType == "application/octet-stream" && downloadType != "" {
		mimeType, _, err = mime.ParseMediaType(downloadType)
	}

	if err != nil {
		contentType := resp.Header.Get("Content-Type")
		http.Error(w, fmt.Sprintf("Failed to parse content type %q: %v", contentType, err), http.StatusBadGateway)
		return
	}

	format, ok := formats.FormatByMimeType(mimeType)
	if !ok {
		httpx.ForwardResponse(w, resp)
		return
	}

	deviceType := device.DetectDevice(r.UserAgent())

	if format == formats.ATOM {
		if err := h.serveAtom(w, r, resp, resolvedURL, deviceType); err != nil {
			reqctx.Logger(r.Context()).Error("Failed to render feed", slog.Any("error", err))
		}
		return
	}

	if err := h.serveFile(w, r, resp, deviceType, format); err != nil {
		reqctx.Logger(r.Context()).Error("Failed to process file", slog.Any("error", err))
	}
}

func (h *FeedHandler) resolveQueryURL(queryURL, searchTerm string) (string, error) {
	parsed, err := url.QueryUnescape(queryURL)
	if err != nil {
		return queryURL, fmt.Errorf("failed to unescape query URL %q: %w", queryURL, err)
	}
	queryURL = parsed

	if searchTerm == "" {
		return queryURL, nil
	}

	escaped := url.QueryEscape(searchTerm)
	repl := strings.NewReplacer("{searchTerms}", escaped, "{searchTerms?}", escaped)

	if strings.Contains(queryURL, "{searchTerms") {
		return repl.Replace(queryURL), nil
	}

	// Fall back to appending the search parameter for non-OpenSearch servers
	u, err := url.Parse(queryURL)
	if err == nil {
		q := u.Query()
		q.Set("query", searchTerm)
		u.RawQuery = q.Encode()
		return u.String(), nil
	}

	return queryURL, nil
}

func (h *FeedHandler) serveAtom(w http.ResponseWriter, r *http.Request, resp *http.Response, feedUrl string, deviceType device.DeviceType) error {
	// Read the body so we can fall back to forwarding it on parse/render errors
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Reset body for downstream parsing
	resp.Body = io.NopCloser(bytes.NewReader(body))

	feed, err := opds.ParseFeed(resp.Body, h.debug)
	if err != nil {
		// Reset again so we can forward the full original response body
		resp.Body = io.NopCloser(bytes.NewReader(body))
		httpx.ForwardResponse(w, resp)
		return nil
	}

	// Intercept OpenSearch descriptor links and resolve them to URL templates
	for i, link := range feed.Links {
		if link.Rel == "search" && link.TypeLink == "application/opensearchdescription+xml" {
			base, err := url.Parse(feedUrl)
			if err == nil {
				if rel, err := url.Parse(link.Href); err == nil {
					osdURL := base.ResolveReference(rel).String()

					var customUserAgent string
					if parsedReqURL, err := url.Parse(osdURL); err == nil {
						for _, feedCfg := range h.feeds {
							if parsedFeedURL, err := url.Parse(feedCfg.Url); err == nil {
								if parsedFeedURL.Hostname() == parsedReqURL.Hostname() && feedCfg.UserAgent != "" {
									customUserAgent = feedCfg.UserAgent
									break
								}
							}
						}
					}

					creds := auth.GetCredentials(osdURL, r, h.feeds, h.s)
					osdResp, err := httpx.Fetch(osdURL, 10, func(req *http.Request) {
						if creds != nil {
							req.SetBasicAuth(creds.Username, creds.Password)
						}
						if customUserAgent != "" {
							req.Header.Set("User-Agent", customUserAgent)
						}
					})

					if err == nil && osdResp.StatusCode >= 200 && osdResp.StatusCode < 300 {
						if tmpl, err := opds.ParseOpenSearchTemplate(osdResp.Body); err == nil && tmpl != "" {
							feed.Links[i].Href = tmpl
							feed.Links[i].TypeLink = "application/atom+xml"
						}
						osdResp.Body.Close()
					} else if err == nil {
						osdResp.Body.Close()
					}
				}
			}
		}
	}

	entryID := r.URL.Query().Get("id")
	if entryID != "" {
		var entry opds.Entry
		for _, e := range feed.Entries {
			if e.ID == entryID {
				entry = e
				break
			}
		}
		if entry.ID == "" {
			http.Error(w, "Entry not found", http.StatusNotFound)
			return nil
		}

		params := view.EntryParams{
			URL:              feedUrl,
			Feed:             feed,
			Entry:            entry,
			DeviceType:       deviceType,
			ConverterManager: h.converters,
		}

		view.Render(w, func(buf io.Writer) error { return view.Entry(buf, params) })
		return nil
	}

	params := view.FeedParams{URL: feedUrl, Feed: feed}
	view.Render(w, func(buf io.Writer) error { return view.Feed(buf, params) })
	return nil
}

func (h *FeedHandler) serveFile(w http.ResponseWriter, r *http.Request, resp *http.Response, deviceType device.DeviceType, inputFormat formats.Format) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log := reqctx.Logger(r.Context())

	filename, err := httpx.ParseFilename(resp)
	if err != nil {
		return err
	}
	log = log.With(slog.String("file", filename))

	converter := h.converters.GetConverterForDevice(deviceType, inputFormat)
	if converter == nil {
		httpx.ForwardResponse(w, resp)
		if filename != "" {
			log.Info("Sent File")
		}
		return nil
	}

	epubFile := filepath.Join(h.outputDir, filename)
	if err := httpx.DownloadToFile(epubFile, resp); err != nil {
		return err
	}
	defer os.Remove(epubFile)

	outputFile, err := converter.Convert(log, epubFile)
	if err != nil {
		return err
	}

	if err := httpx.SendFile(w, outputFile, filepath.Base(outputFile)); err != nil {
		return err
	}

	log.Info("Sent Converted File", slog.String("converter", reflect.TypeOf(converter).String()))
	return nil
}
