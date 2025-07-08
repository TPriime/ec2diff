package pkg

const (
	CommentDriftDetected   = "Drifts detected"
	CommentNoDriftDetected = "No drifts detected"
	CommentMissingState    = "Missing state"
)

// ReportPrinter defines how reports would be printed.
type ReportPrinter interface {
	Print(reports []Report)
}

// Report captures drift for one instance
type Report struct {
	InstanceID string           `json:"instance_id"`
	Drifts     []AttributeDrift `json:"drifts"`
	Comment    string           `json:"comment"`
}

// AttributeDrift describes an attribute mismatch
type AttributeDrift struct {
	Name     string `json:"name"`
	Expected any    `json:"expected"`
	Found    any    `json:"found"`
}
