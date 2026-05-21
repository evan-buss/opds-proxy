package formats

import (
	"testing"
)

func TestFormatByMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     Format
		wantOk   bool
	}{
		{
			name:     "ATOM from standard application/atom+xml",
			mimeType: "application/atom+xml",
			want:     ATOM,
			wantOk:   true,
		},
		{
			name:     "ATOM from generic application/xml",
			mimeType: "application/xml",
			want:     ATOM,
			wantOk:   true,
		},
		{
			name:     "ATOM from generic text/xml",
			mimeType: "text/xml",
			want:     ATOM,
			wantOk:   true,
		},
		{
			name:     "EPUB from standard application/epub+zip",
			mimeType: "application/epub+zip",
			want:     EPUB,
			wantOk:   true,
		},
		{
			name:     "Unknown format",
			mimeType: "application/unknown",
			want:     Format{},
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FormatByMimeType(tt.mimeType)
			if ok != tt.wantOk {
				t.Errorf("FormatByMimeType() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if ok && got != tt.want {
				t.Errorf("FormatByMimeType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
