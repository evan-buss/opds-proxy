package opds

import (
	"encoding/xml"
	"fmt"
	"io"
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

// ParseOpenSearchTemplate parses an OpenSearch Description XML from r and
// returns the preferred Atom/OPDS search template URL.
func ParseOpenSearchTemplate(r io.Reader) (string, error) {
	// Stream decode to avoid buffering entire body in memory
	var d OpenSearchDescription
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&d); err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to parse OpenSearch description: %w", err)
	}

	// First pass: look for the OPDS profile type with kind=acquisition
	for _, u := range d.Urls {
		if u.Type == "application/atom+xml;profile=opds-catalog;kind=acquisition" && u.Template != "" {
			return u.Template, nil
		}
	}
	// Second pass: look for the OPDS profile type
	for _, u := range d.Urls {
		if u.Type == "application/atom+xml;profile=opds-catalog" && u.Template != "" {
			return u.Template, nil
		}
	}
	// Third pass: any atom+xml template
	for _, u := range d.Urls {
		if u.Type == "application/atom+xml" && u.Template != "" {
			return u.Template, nil
		}
	}
	// Fourth pass: any template at all
	for _, u := range d.Urls {
		if u.Template != "" {
			return u.Template, nil
		}
	}

	return "", fmt.Errorf("no suitable template found in OSDD")
}
