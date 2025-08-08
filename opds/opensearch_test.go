package opds

import (
	"strings"
	"testing"
)

func TestParseOpenSearchTemplate_PrefersOPDSProfile(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription>
  <Url type="application/atom+xml" template="https://example.test/search?q={searchTerms}"/>
  <Url type="application/atom+xml;profile=opds-catalog" template="https://example.test/opds?q={searchTerms}"/>
</OpenSearchDescription>`

	tmpl, err := parseOpenSearchTemplate(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "https://example.test/opds?q={searchTerms}"
	if tmpl != want {
		t.Fatalf("unexpected template. got %q want %q", tmpl, want)
	}
}

func TestParseOpenSearchTemplate_FallbackAtom(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription>
  <Url type="application/atom+xml" template="https://example.test/search?q={searchTerms}"/>
</OpenSearchDescription>`

	tmpl, err := parseOpenSearchTemplate(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "https://example.test/search?q={searchTerms}"
	if tmpl != want {
		t.Fatalf("unexpected template. got %q want %q", tmpl, want)
	}
}

func TestParseOpenSearchTemplate_NoSuitable(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription>
  <Url type="text/html" template="https://example.test/html?q={searchTerms}"/>
  <Url type="application/json" template="https://example.test/api?q={searchTerms}"/>
</OpenSearchDescription>`

	_, err := parseOpenSearchTemplate(strings.NewReader(xml))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestParseOpenSearchTemplate_EmptyTemplateIgnored(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription>
  <Url type="application/atom+xml;profile=opds-catalog" template=""/>
  <Url type="application/atom+xml" template="https://example.test/search?q={searchTerms}"/>
</OpenSearchDescription>`

	tmpl, err := parseOpenSearchTemplate(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "https://example.test/search?q={searchTerms}"
	if tmpl != want {
		t.Fatalf("unexpected template. got %q want %q", tmpl, want)
	}
}

func TestParseOpenSearchTemplate_MalformedXML(t *testing.T) {
	// Missing closing tag
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription>
  <Url type="application/atom+xml" template="https://example.test/search?q={searchTerms}"/>`

	_, err := parseOpenSearchTemplate(strings.NewReader(xml))
	if err == nil {
		t.Fatalf("expected XML parse error, got nil")
	}
}
