package httpx

import (
	"fmt"
	"io"
	"maps"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Fetch(url string, timeoutSeconds int, setAuth func(*http.Request)) (*http.Response, error) {
	client := &http.Client{}
	if timeoutSeconds > 0 {
		client.Timeout = time.Duration(timeoutSeconds) * time.Second
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if setAuth != nil {
		setAuth(req)
	}
	return client.Do(req)
}

func ForwardResponse(w http.ResponseWriter, resp *http.Response) {
	maps.Copy(w.Header(), resp.Header)
	io.Copy(w, resp.Body)
}

func ParseFilename(resp *http.Response) (string, error) {
	contentDisposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err == nil && params["filename"] != "" {
		return params["filename"], nil
	}
	if parsedUrl, err := url.Parse(resp.Request.URL.String()); err == nil {
		return path.Base(parsedUrl.Path), nil
	}
	return "", err
}

func DownloadToFile(dstPath string, resp *http.Response) error {
	file, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func SendFile(w http.ResponseWriter, filePath, outFilename string) error {
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
			map[string]string{"filename": sanitizeFilenameASCII7(outFilename)},
		),
	)
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filePath)))

	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}
	return nil
}

func sanitizeFilenameASCII7(s string) string {
	// Remove most diacritics and nonspacing marks (Mn)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	noDiacr, _, _ := transform.String(t, s)

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

// WriteRecorder copies the contents of a ResponseRecorder to a ResponseWriter
func WriteRecorder(rec *httptest.ResponseRecorder, w http.ResponseWriter) {
	maps.Copy(w.Header(), rec.Header())
	w.WriteHeader(rec.Code)
	_, _ = w.Write(rec.Body.Bytes())
}
