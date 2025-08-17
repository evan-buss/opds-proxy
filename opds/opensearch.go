package opds

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenSearchDescription represents an OpenSearch Description Document (OSDD)
// per https://specs.opds.io/opds-1.2#opensearch
// We only need the <Url> elements with a template.
type OpenSearchDescription struct {
	XMLName xml.Name `xml:"OpenSearchDescription"`
	Urls    []OSDUrl `xml:"Url"`
}

// OSDUrl represents a single <Url> entry in an OSDD
// Example:
// <Url type="application/atom+xml;profile=opds-catalog" template="https://example.org/search?q={searchTerms}"/>
// Some servers might omit the profile and use just application/atom+xml.
type OSDUrl struct {
	Type     string `xml:"type,attr"`
	Template string `xml:"template,attr"`
}

// ResolveOpenSearchTemplate fetches an OSDD from the given URL and returns the
// Atom/OPDS template URL to use for search requests. It prefers
// "application/atom+xml;profile=opds-catalog" then falls back to
// "application/atom+xml" if needed.
func ResolveOpenSearchTemplate(osdURL string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// simplified request: use client.Get since no headers are needed
	resp, err := client.Get(osdURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch OpenSearch description from %q: %w", osdURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status fetching OSDD: %s", resp.Status)
	}

	return parseOpenSearchTemplate(resp.Body)
}

// parseOpenSearchTemplate parses an OpenSearch Description XML from r and
// returns the preferred Atom/OPDS search template URL.
func parseOpenSearchTemplate(r io.Reader) (string, error) {
	// Stream decode to avoid buffering entire body in memory
	var d OpenSearchDescription
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&d); err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to parse OpenSearch description: %w", err)
	}

	// First pass: look for the OPDS profile type
	for _, u := range d.Urls {
		if u.Type == "application/atom+xml;profile=opds-catalog" && u.Template != "" {
			return u.Template, nil
		}
	}
	// Second pass: any atom+xml template
	for _, u := range d.Urls {
		if u.Type == "application/atom+xml" && u.Template != "" {
			return u.Template, nil
		}
	}

	return "", fmt.Errorf("no suitable Atom template found in OSDD")
}
