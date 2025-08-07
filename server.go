package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/evan-buss/opds-proxy/convert"
	"github.com/evan-buss/opds-proxy/html"
	"github.com/evan-buss/opds-proxy/internal/debounce"
	"github.com/evan-buss/opds-proxy/internal/device"
	"github.com/evan-buss/opds-proxy/internal/formats"
	"github.com/evan-buss/opds-proxy/opds"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
)

var (
	_ = mime.AddExtensionType(formats.EPUB.Extension, formats.EPUB.MimeType)
	_ = mime.AddExtensionType(formats.KEPUB.Extension, formats.KEPUB.MimeType)
	_ = mime.AddExtensionType(formats.MOBI.Extension, formats.MOBI.MimeType)
)

type Server struct {
	addr   string
	router *http.ServeMux
	s      *securecookie.SecureCookie
}

type Credentials struct {
	Username string
	Password string
}

type contextKey string

const (
	requestLogger  = contextKey("requestLogger")
	isLocalRequest = contextKey("isLocalRequest")
)

const cookieName = "auth-creds"

func NewServer(config *ProxyConfig) (*Server, error) {
	hashKey, err := hex.DecodeString(config.Auth.HashKey)
	if err != nil {
		return nil, err
	}
	blockKey, err := hex.DecodeString(config.Auth.BlockKey)
	if err != nil {
		return nil, err
	}

	if !config.DebugMode {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

	s := securecookie.New(hashKey, blockKey)

	// Kobo issues 2 requests for each clicked link. This middleware ensures
	// we only process the first request and provide the same response for the second.
	// This becomes more important when the requests aren't idempotent, such as triggering
	// a download.
	debounceMiddleware := debounce.NewDebounceMiddleware(time.Millisecond * 100)

	router := http.NewServeMux()
	router.Handle("GET /{$}", requestMiddleware(handleHome(config.Feeds)))
	router.Handle("GET /feed", requestMiddleware(debounceMiddleware(handleFeed("tmp/", config.Feeds, s, config.DebugMode))))
	router.Handle("/auth", requestMiddleware(handleAuth(s)))
	router.Handle("GET /static/", http.FileServer(http.FS(html.StaticFiles())))

	return &Server{
		addr:   ":" + config.Port,
		router: router,
		s:      s,
	}, nil
}

func requestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := uuid.New()
		requestIP := r.Header.Get("X-Forwarded-For")
		if requestIP == "" {
			requestIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		}

		isLocal := true
		for addr := range strings.SplitSeq(requestIP, ", ") {
			ip := net.ParseIP(addr)
			if ip == nil || (!ip.IsPrivate() && !ip.IsLoopback()) {
				isLocal = false
				break
			}
		}
		ctx := context.WithValue(r.Context(), isLocalRequest, isLocal)

		query, _ := url.QueryUnescape(r.URL.RawQuery)
		log := slog.With(
			slog.Group("request",
				slog.String("id", id.String()),
				slog.String("ip", requestIP),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", query),
				slog.String("user-agent", r.UserAgent()),
			),
		)

		ctx = context.WithValue(ctx, requestLogger, log)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		log.Info("Request Completed",
			slog.String("duration", time.Since(start).String()),
			slog.Bool("debounce", w.Header().Get("X-Debounce") == "true"),
			slog.Bool("shared", w.Header().Get("X-Shared") == "true"),
		)
	})
}

func (s *Server) Serve() error {
	slog.Info("Starting server", slog.String("port", s.addr))
	return http.ListenAndServe(s.addr, s.router)
}

func handleHome(feeds []FeedConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Don't make user click the only feed
		if len(feeds) == 1 {
			http.Redirect(w, r, "/feed?q="+feeds[0].Url, http.StatusFound)
			return
		}

		homeParams := make([]html.HomeParams, len(feeds))
		for i, feed := range feeds {
			homeParams[i] = html.HomeParams{
				Title: feed.Name,
				URL:   feed.Url,
			}
		}

		html.Home(w, homeParams)
	}
}

func handleFeed(outputDir string, feeds []FeedConfig, s *securecookie.SecureCookie, debug bool) http.HandlerFunc {
	converterManager := convert.NewConverterManager()

	mutex := sync.Mutex{}

	return func(w http.ResponseWriter, r *http.Request) {
		queryURL := r.URL.Query().Get("q")
		if queryURL == "" {
			http.Error(w, "No feed specified", http.StatusBadRequest)
			return
		}

		parsedUrl, err := url.QueryUnescape(queryURL)
		queryURL = parsedUrl
		if err != nil {
			handleError(r, w, "Failed to parse URL", err)
			return
		}

		searchTerm := r.URL.Query().Get("search")
		if searchTerm != "" {
			queryURL = strings.Replace(queryURL, "{searchTerms}", searchTerm, 1)
		}

		resp, err := fetchFromUrl(queryURL, getCredentials(queryURL, r, feeds, s))
		if err != nil {
			handleError(r, w, "Failed to fetch", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			http.Redirect(w, r, "/auth?return="+r.URL.String(), http.StatusFound)
			return
		}

		mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			handleError(r, w, "Failed to parse content type", err)
			return
		}

		format, exists := formats.FormatByMimeType(mimeType)
		if !exists {
			forwardResponse(w, resp)
			return
		}

		deviceType := device.DetectDevice(r.UserAgent())

		if format == formats.ATOM {
			feed, err := opds.ParseFeed(resp.Body, debug)
			if err != nil {
				handleError(r, w, "Failed to parse feed", err)
				return
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
					handleError(r, w, "Entry not found", nil)
					return
				}

				entryParams := html.EntryParams{
					URL:              queryURL,
					Feed:             feed,
					Entry:            entry,
					DeviceType:       deviceType,
					ConverterManager: converterManager,
				}

				if err := html.Entry(w, entryParams); err != nil {
					handleError(r, w, "Failed to render entry", err)
				}

				return
			}

			feedParams := html.FeedParams{
				URL:  queryURL,
				Feed: feed,
			}

			if err = html.Feed(w, feedParams); err != nil {
				handleError(r, w, "Failed to render feed", err)
				return
			}
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		log := r.Context().Value(requestLogger).(*slog.Logger)
		filename, err := parseFileName(resp)
		if err != nil {
			handleError(r, w, "Failed to parse file name", err)
			return
		}

		log = log.With(slog.String("file", filename))

		converter := converterManager.GetConverterForDevice(deviceType, format)

		if converter == nil {
			forwardResponse(w, resp)
			if filename != "" {
				log.Info("Sent File")
			}
			return
		}

		epubFile := filepath.Join(outputDir, filename)
		downloadFile(epubFile, resp)
		defer os.Remove(epubFile)

		outputFile, err := converter.Convert(log, epubFile)
		if err != nil {
			handleError(r, w, "Failed to convert epub", err)
			return
		}

		if err = sendConvertedFile(w, outputFile); err != nil {
			handleError(r, w, "Failed to send converted file", err)
			return
		}

		log.Info("Sent Converted File", slog.String("converter", reflect.TypeOf(converter).String()))
	}
}

func handleAuth(s *securecookie.SecureCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		returnUrl := r.URL.Query().Get("return")
		if returnUrl == "" {
			http.Error(w, "No return URL specified", http.StatusBadRequest)
			return
		}

		if r.Method == "GET" {
			html.Login(w, html.LoginParams{ReturnURL: returnUrl})
			return
		}

		if r.Method == "POST" {
			username := r.FormValue("username")
			password := r.FormValue("password")

			rUrl, err := url.Parse(returnUrl)
			if err != nil {
				http.Error(w, "Invalid return URL", http.StatusBadRequest)
			}
			domain, err := url.Parse(rUrl.Query().Get("q"))
			if err != nil {
				http.Error(w, "Invalid site", http.StatusBadRequest)
			}

			value := map[string]Credentials{
				domain.Hostname(): {Username: username, Password: password},
			}

			encoded, err := s.Encode(cookieName, value)
			if err != nil {
				handleError(r, w, "Failed to encode credentials", err)
				return
			}
			cookie := &http.Cookie{
				Name:  cookieName,
				Value: encoded,
				Path:  "/",
				// Kobo fails to set cookies with HttpOnly or Secure flags
				Secure:   false,
				HttpOnly: false,
			}

			http.SetCookie(w, cookie)
			http.Redirect(w, r, returnUrl, http.StatusFound)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getCredentials(rawUrl string, req *http.Request, feeds []FeedConfig, s *securecookie.SecureCookie) *Credentials {
	requestUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil
	}

	// Try to get credentials from the config first
	for _, feed := range feeds {
		feedUrl, err := url.Parse(feed.Url)
		if err != nil {
			continue
		}

		if feedUrl.Hostname() != requestUrl.Hostname() {
			continue
		}

		if feed.Auth == nil || feed.Auth.Username == "" || feed.Auth.Password == "" {
			continue
		}

		// Only set feed credentials for local requests
		// when the auth config has local_only flag
		isLocal := req.Context().Value(isLocalRequest).(bool)
		if !isLocal && feed.Auth.LocalOnly {
			continue
		}

		return &Credentials{Username: feed.Auth.Username, Password: feed.Auth.Password}
	}

	// Otherwise, try to get credentials from the cookie
	cookie, err := req.Cookie(cookieName)
	if err != nil {
		return nil
	}

	value := make(map[string]*Credentials)
	if err = s.Decode(cookieName, cookie.Value, &value); err != nil {
		return nil
	}

	return value[requestUrl.Hostname()]
}

func fetchFromUrl(url string, credentials *Credentials) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}

	return client.Do(req)
}

func handleError(r *http.Request, w http.ResponseWriter, message string, err error) {
	log := r.Context().Value(requestLogger).(*slog.Logger)
	log.Error(message, slog.Any("error", err))
	http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
}

func downloadFile(path string, resp *http.Response) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func parseFileName(resp *http.Response) (string, error) {
	contentDisposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err == nil && params["filename"] != "" {
		return params["filename"], nil
	}

	// Fallback to using the last part of the URL as the filename
	if parsedUrl, err := url.Parse(resp.Request.URL.String()); err == nil {
		return path.Base(parsedUrl.Path), nil
	}

	return "", err
}

func forwardResponse(w http.ResponseWriter, resp *http.Response) {
	maps.Copy(w.Header(), resp.Header)
	io.Copy(w, resp.Body)
}

func sendConvertedFile(w http.ResponseWriter, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		os.Remove(filePath)
		return err
	}
	defer func() {
		file.Close()
		os.Remove(filePath)
	}()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	w.Header().Set("Content-Disposition",
		mime.FormatMediaType(
			"attachment",
			map[string]string{"filename": filenameToAscii7(filepath.Base(filePath))},
		),
	)
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filePath)))

	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}

	return nil
}

func filenameToAscii7(s string) string {
	// Remove most diacritics and nonspacing marks (Mn)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	noDiacr, _, _ := transform.String(t, s)

	// Convert the rest of non-ASCII7 to hex representation
	var sb strings.Builder
	for _, letter := range noDiacr {
		if letter > unicode.MaxASCII {
			sb.WriteString(fmt.Sprintf("_%X", letter))
		} else {
			sb.WriteRune(letter)
		}
	}

	return sb.String()
}
