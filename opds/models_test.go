package opds

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestFlexibleTimeUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		xmlData  string
		expected string
	}{
		{
			name:     "RFC3339 format",
			xmlData:  `<wrapper><published>2006-01-02T15:04:05Z</published></wrapper>`,
			expected: "2006-01-02T15:04:05Z",
		},
		{
			name:     "Date only format",
			xmlData:  `<wrapper><published>2006-07-18</published></wrapper>`,
			expected: "2006-07-18T00:00:00Z",
		},
		{
			name:     "Date with time no timezone",
			xmlData:  `<wrapper><published>2006-01-02T15:04:05</published></wrapper>`,
			expected: "2006-01-02T15:04:05Z",
		},
		{
			name:     "Empty published date",
			xmlData:  `<wrapper><published></published></wrapper>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				Published *FlexibleTime `xml:"published"`
			}

			err := xml.NewDecoder(strings.NewReader(tt.xmlData)).Decode(&result)
			if err != nil {
				t.Fatalf("Failed to unmarshal XML: %v", err)
			}

			if tt.expected == "" {
				if result.Published != nil && !result.Published.Time.IsZero() {
					t.Errorf("Expected nil or zero time, got %v", result.Published.Time)
				}
				return
			}

			if result.Published == nil {
				t.Errorf("Expected non-nil published time")
				return
			}

			actual := result.Published.Time.UTC().Format(time.RFC3339)
			if actual != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestEntryWithDateOnlyPublished(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<entry xmlns="http://www.w3.org/2005/Atom">
	<title>Test Book</title>
	<id>test-id</id>
	<published>2006-07-18</published>
</entry>`

	var entry Entry
	err := xml.NewDecoder(strings.NewReader(xmlData)).Decode(&entry)
	if err != nil {
		t.Fatalf("Failed to unmarshal entry: %v", err)
	}

	if entry.Published == nil {
		t.Error("Expected published date to be parsed")
		return
	}

	expected := time.Date(2006, 7, 18, 0, 0, 0, 0, time.UTC)
	if !entry.Published.Time.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, entry.Published.Time)
	}
}
