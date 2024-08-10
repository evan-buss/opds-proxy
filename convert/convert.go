package convert

import "log/slog"

type Converter interface {
	// Whether or not the converter is available
	// Usually based on the availability of the underlying tool
	Available() bool
	// Convert the input file to the output file
	Convert(log *slog.Logger, input string) (string, error)
}
