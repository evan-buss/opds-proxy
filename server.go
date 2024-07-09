package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/evan-buss/kobo-opds-proxy/html"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Server struct {
	addr   string
	router *http.ServeMux
}

func NewServer() *Server {
	port := os.Getenv("PORT")
	if os.Getenv("PORT") == "" {
		port = "8080"
	}

	fmt.Println(os.Getenv("PATH"))

	router := http.NewServeMux()
	router.HandleFunc("GET /{$}", handleHome())
	router.HandleFunc("GET /feed", handleFeed("tmp/"))
	router.Handle("GET /static/", http.FileServer(http.FS(html.StaticFiles())))

	return &Server{
		addr:   ":" + port,
		router: router,
	}
}
