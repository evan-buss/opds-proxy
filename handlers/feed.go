package handlers

import (
	"bytes"
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
	if queryURL == "" {
		http.Error(w, "No feed specified", http.StatusBadRequest)
		return
	}

	resolvedURL, err := h.resolveQueryURL(queryURL, r.URL.Query().Get("search"))
	if err != nil {
		http.Error(w, "Failed to parse URL", http.StatusBadRequest)
		return
	}

	creds := auth.GetCredentials(resolvedURL, r, h.feeds, h.s)
	resp, err := httpx.Fetch(resolvedURL, 10, func(req *http.Request) {
		if creds != nil {
			req.SetBasicAuth(creds.Username, creds.Password)
		}
	})
	if err != nil {
		http.Error(w, "Failed to fetch", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		http.Redirect(w, r, "/auth?return="+r.URL.String(), http.StatusFound)
		return
	}

	mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "Failed to parse content type", http.StatusBadGateway)
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
		return queryURL, err
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

	if tmpl, err := opds.ResolveOpenSearchTemplate(queryURL); err == nil && tmpl != "" {
		return repl.Replace(tmpl), nil
	}

	return queryURL, nil
}

func (h *FeedHandler) serveAtom(w http.ResponseWriter, r *http.Request, resp *http.Response, url string, deviceType device.DeviceType) error {
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
			URL:              url,
			Feed:             feed,
			Entry:            entry,
			DeviceType:       deviceType,
			ConverterManager: h.converters,
		}

		view.Render(w, func(buf io.Writer) error { return view.Entry(buf, params) })
		return nil
	}

	params := view.FeedParams{URL: url, Feed: feed}
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
