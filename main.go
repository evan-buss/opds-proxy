package main

import (
	"encoding/xml"
	"github.com/opds-community/libopds2-go/opds1"
	"log"
	"net/http"
	"os"
	"github.com/evan-buss/kobo-opds-proxy/html"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		feed := parseFeed()
		html.Feed(w, html.FeedParams{feed}, "")
	})

	router.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	log.Panic(http.ListenAndServe(":8080", router))
}

func parseFeed() *opds1.Feed {
	var feed opds1.Feed

	bytes, err := os.ReadFile("/home/evan/Downloads/opds.atom")
	if err != nil {
		panic(err)
	}
	err = xml.Unmarshal(bytes, &feed)
	if err != nil {
		panic(err)
	}

	return &feed
}
