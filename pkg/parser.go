package pkg

// Parser defines how to extract instances from supported config files (e.g. HCL, JSON).
type Parser interface {

	// Parse reads the given file and returns instances by ID.
	//
	// filePath: path to the input file
	Parse(filePath string) (InstanceMap, error)

	// SupportedTypes lists the file extensions this parser can handle.
	SupportedTypes() []string
}
