package opds

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ParseFeed parses an OPDS feed from an io.Reader and returns a Feed struct.
// It also writes the raw XML to a temporary file for debugging purposes.
func ParseFeed(r io.Reader, debug bool) (*Feed, error) {
	var feed Feed

	if !debug {
		err := xml.NewDecoder(r).Decode(&feed)
		if err != nil {
			return nil, err
		}
		return &feed, nil
	}

	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Write raw XML to temp directory
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("opds_feed_%s.xml", timestamp)
	filepath := filepath.Join(tempDir, filename)

	err = os.WriteFile(filepath, body, 0644)
	if err != nil {
		slog.Error("Failed to write raw OPDS feed to file", slog.Any("error", err))
	} else {
		slog.Debug("Raw OPDS feed written to file", slog.String("filepath", filepath))
	}

	err = xml.NewDecoder(strings.NewReader(string(body))).Decode(&feed)
	if err != nil {
		return nil, err
	}

	return &feed, nil
}
