package main

import (
	"context"
	"encoding/hex"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/evan-buss/opds-proxy/handlers"
	"github.com/evan-buss/opds-proxy/internal/auth"
	"github.com/evan-buss/opds-proxy/internal/debounce"
	"github.com/evan-buss/opds-proxy/internal/formats"
	"github.com/evan-buss/opds-proxy/internal/reqctx"
	"github.com/evan-buss/opds-proxy/view"
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

func NewServer(configData *ProxyConfig) (*Server, error) {
	hashKey, err := hex.DecodeString(configData.Auth.HashKey)
	if err != nil {
		return nil, err
	}
	blockKey, err := hex.DecodeString(configData.Auth.BlockKey)
	if err != nil {
		return nil, err
	}

	if !configData.DebugMode {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

	s := securecookie.New(hashKey, blockKey)

	// Kobo issues 2 requests for each clicked link. This middleware ensures
	// we only process the first request and provide the same response for the second.
	// This becomes more important when the requests aren't idempotent, such as triggering
	// a download.
	debounceMiddleware := debounce.NewDebounceMiddleware(time.Millisecond * 100)

	router := http.NewServeMux()
	// Home
	links := make([]handlers.HomeLink, len(configData.Feeds))
	for i, f := range configData.Feeds {
		links[i] = handlers.HomeLink{Title: f.Name, URL: f.Url}
	}
	router.Handle("GET /{$}", requestMiddleware(handlers.Home(links)))

	// Feed
	adapted := make([]auth.FeedConfig, len(configData.Feeds))
	for i, f := range configData.Feeds {
		adapted[i] = auth.FeedConfig{Name: f.Name, Url: f.Url, Auth: toAuthPtr(f.Auth)}
	}
	router.Handle("GET /feed", requestMiddleware(debounceMiddleware(handlers.Feed("tmp/", adapted, s, configData.DebugMode))))

	// Auth
	router.Handle("/auth", requestMiddleware(handlers.Auth(s)))

	// Static assets (serve embedded files from view package)
	router.Handle("GET /static/", http.FileServer(http.FS(view.StaticFiles())))

	return &Server{addr: ":" + configData.Port, router: router, s: s}, nil
}

func toAuthPtr(a *FeedConfigAuth) *auth.FeedAuth {
	if a == nil {
		return nil
	}
	return &auth.FeedAuth{Username: a.Username, Password: a.Password, LocalOnly: a.LocalOnly}
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

		ctx := reqctx.WithIsLocal(context.Background(), isLocal)
		ctx = reqctx.WithRequestLogger(ctx, log)
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
