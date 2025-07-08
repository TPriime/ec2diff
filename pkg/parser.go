package pkg

type InstanceMap = map[string]Instance

// Parser defines an interface for extracting instance data from supported infrastructure-as-code files.
// It enables pluggable parsing strategies—e.g., for HCL, JSON, or other formats.
type Parser interface {

	// Parse reads the file at the given path and extracts instance definitions.
	//
	// Parameters:
	// - filePath: path to the file to parse (e.g., .hcl, .tf)
	// - ids: list of instance IDs to extract; mapping is done based on index or external logic
	//
	// Returns:
	// - map of instance ID to corresponding Instance struct
	// - error if parsing fails or file does not conform to expected structure
	Parse(filePath string, ids []string) (InstanceMap, error)

	// SupportedTypes returns a list of file extensions this parser supports.
	//
	// This allows a dispatcher or factory method to select the appropriate parser
	// implementation based on the input file type.
	SupportedTypes() []string
}
