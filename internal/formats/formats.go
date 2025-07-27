package formats

// Format represents a supported ebook format
type Format struct {
	// MIME type for the format
	MimeType string
	// File extension (including the dot)
	Extension string
	// Human-readable label
	Label string
	// Whether this format can be converted from EPUB
	ConvertibleFromEPUB bool
}

// Supported ebook formats
var (
	EPUB = Format{
		MimeType:            "application/epub+zip",
		Extension:           ".epub",
		Label:               "EPUB",
		ConvertibleFromEPUB: false, // Source format
	}

	KEPUB = Format{
		MimeType:            "application/epub+zip", // KEPUB uses same MIME as EPUB
		Extension:           ".kepub.epub",
		Label:               "KEPUB",
		ConvertibleFromEPUB: true,
	}

	MOBI = Format{
		MimeType:            "application/x-mobipocket-ebook",
		Extension:           ".mobi",
		Label:               "MOBI",
		ConvertibleFromEPUB: true,
	}

	PDF = Format{
		MimeType:            "application/pdf",
		Extension:           ".pdf",
		Label:               "PDF",
		ConvertibleFromEPUB: false, // Not converted, just served
	}

	AZW3 = Format{
		MimeType:            "application/x-mobi8-ebook",
		Extension:           ".azw3",
		Label:               "AZW3",
		ConvertibleFromEPUB: false, // Not currently supported for conversion
	}

	// OPDS/Atom feed format
	ATOM = Format{
		MimeType:            "application/atom+xml",
		Extension:           ".xml",
		Label:               "ATOM",
		ConvertibleFromEPUB: false,
	}
)

// AllFormats returns all supported formats
func AllFormats() []Format {
	return []Format{EPUB, KEPUB, MOBI, PDF, AZW3, ATOM}
}

// FormatByMimeType returns the format for a given MIME type
func FormatByMimeType(mimeType string) (Format, bool) {
	formats := map[string]Format{
		EPUB.MimeType:  EPUB,
		KEPUB.MimeType: EPUB, // KEPUB shares MIME with EPUB
		MOBI.MimeType:  MOBI,
		PDF.MimeType:   PDF,
		AZW3.MimeType:  AZW3,
		ATOM.MimeType:  ATOM,
		// Legacy/alternative MIME types
		"application/mobi":       MOBI,
		"application/x-epub+zip": EPUB,
	}
	
	format, exists := formats[mimeType]
	return format, exists
}

// FormatByExtension returns the format for a given file extension
func FormatByExtension(extension string) (Format, bool) {
	formats := map[string]Format{
		EPUB.Extension:  EPUB,
		KEPUB.Extension: KEPUB,
		MOBI.Extension:  MOBI,
		PDF.Extension:   PDF,
		AZW3.Extension:  AZW3,
		ATOM.Extension:  ATOM,
	}
	
	format, exists := formats[extension]
	return format, exists
}

// GetMimeTypeLabel returns the human-readable label for a MIME type
func GetMimeTypeLabel(mimeType string) string {
	if format, exists := FormatByMimeType(mimeType); exists {
		return format.Label
	}
	return "Unknown"
}

// ConvertibleFormats returns all formats that can be converted from EPUB
func ConvertibleFormats() []Format {
	var convertible []Format
	for _, format := range AllFormats() {
		if format.ConvertibleFromEPUB {
			convertible = append(convertible, format)
		}
	}
	return convertible
}