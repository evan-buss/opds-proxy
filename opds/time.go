package opds

import (
	"encoding/xml"
	"strings"
	"time"
)

// Time handles multiple time formats commonly found in OPDS feeds
type Time struct {
	time.Time
}

// UnmarshalXML implements xml.Unmarshaler to handle multiple date formats
func (ft *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}

	// Remove any extra whitespace
	v = strings.TrimSpace(v)
	if v == "" {
		ft.Time = time.Time{}
		return nil
	}

	// Try different time formats commonly used in OPDS feeds
	formats := []string{
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,       // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05Z", // Without timezone offset
		"2006-01-02T15:04:05",  // Without timezone
		"2006-01-02",           // Date only
	}

	for _, format := range formats {
		if t, err := time.Parse(format, v); err == nil {
			ft.Time = t
			return nil
		}
	}

	// If none of the formats work, try to parse just the date part
	if len(v) >= 10 {
		if t, err := time.Parse("2006-01-02", v[:10]); err == nil {
			ft.Time = t
			return nil
		}
	}

	// Return a zero time if we can't parse it
	ft.Time = time.Time{}
	return nil
}

// MarshalXML implements xml.Marshaler to output in RFC3339 format
func (ft Time) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if ft.Time.IsZero() {
		return nil
	}
	return e.EncodeElement(ft.Time.Format(time.RFC3339), start)
}
