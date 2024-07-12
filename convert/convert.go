package convert

type Converter interface {
	// Whether or not the converter is available
	// Usually based on the availability of the underlying tool
	Available() bool
	// Convert the input file to the output file
	Convert(input string) (string, error)
}
