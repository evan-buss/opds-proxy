package convert

import (
	"log/slog"

	"github.com/evan-buss/opds-proxy/internal/formats"
)

type Converter interface {
	// Whether or not the converter is available
	// Usually based on the availability of the underlying tool
	Available() bool
	// Determine if the convert handles the input format
	HandlesInputFormat(format formats.Format) bool
	// Convert the input file to the output file
	Convert(log *slog.Logger, input string) (string, error)
}
