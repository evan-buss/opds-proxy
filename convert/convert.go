package convert

const (
	MOBI_MIME = "application/x-mobipocket-ebook"
	EPUB_MIME = "application/epub+zip"
)

type Converter interface {
	// Whether or not the converter is available
	// Usually based on the availability of the underlying tool
	Available() bool
	// Convert the input file to the output file
	Convert(input string, output string) error
}
