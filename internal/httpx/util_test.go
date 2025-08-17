package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteRecorder(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Test", "1")
	rec.WriteHeader(http.StatusTeapot)
	_, _ = rec.Write([]byte("hello"))

	rw := httptest.NewRecorder()
	WriteRecorder(rec, rw)

	if rw.Code != http.StatusTeapot {
		t.Fatalf("unexpected code: %d", rw.Code)
	}
	if got := rw.Header().Get("X-Test"); got != "1" {
		t.Fatalf("unexpected header: %q", got)
	}
	if got := rw.Body.String(); got != "hello" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestSanitizeFilenameASCII7(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"Resume.txt", "Resume.txt"},
		{"Résumé.txt", "Resume.txt"},
		{"Cafe\u0301.txt", "Cafe.txt"}, // decomposed acute accent
		{"你好.kepub.epub", "_4F60_597D.kepub.epub"},
		{"Café – 你好?.epub", "Cafe _2013 _4F60_597D?.epub"},
		{"Hello-World_123.txt", "Hello-World_123.txt"},
		{"file-📚.txt", "file-_1F4DA.txt"},
	}
	for _, c := range cases {
		if got := sanitizeFilenameASCII7(c.in); got != c.out {
			t.Fatalf("sanitizeFilenameASCII7(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestParseFilename_ContentDisposition(t *testing.T) {
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("Content-Disposition", "attachment; filename=\"report.epub\"")
	name, err := ParseFilename(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "report.epub" {
		t.Fatalf("got %q, want %q", name, "report.epub")
	}
}

func TestParseFilename_FromURL(t *testing.T) {
	u, _ := url.Parse("https://example.com/files/book.epub?x=1")
	resp := &http.Response{Request: &http.Request{URL: u}}
	name, err := ParseFilename(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "book.epub" {
		t.Fatalf("got %q, want %q", name, "book.epub")
	}
}

func TestDownloadToFile(t *testing.T) {
	tmp, err := os.MkdirTemp("", "httpx-dl-*")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	defer os.RemoveAll(tmp)

	out := filepath.Join(tmp, "out.bin")
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("hello world"))}
	defer resp.Body.Close()

	if err := DownloadToFile(out, resp); err != nil {
		t.Fatalf("DownloadToFile error: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	if string(b) != "hello world" {
		t.Fatalf("unexpected content: %q", string(b))
	}
}
