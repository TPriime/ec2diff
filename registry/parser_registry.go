package registry

import (
	"path/filepath"

	"github.com/tpriime/ec2diff/pkg"
)

// ParserRegistry maps file extensions to parser implementations.
type ParserRegistry struct {
	parsers map[string]pkg.Parser
}

// NewParserRegistry constructs a registry from a list of parsers.
func NewParserRegistry(parsers []pkg.Parser) *ParserRegistry {
	reg := &ParserRegistry{parsers: make(map[string]pkg.Parser)}
	for _, p := range parsers {
		for _, ext := range p.SupportedTypes() {
			reg.parsers[ext] = p
		}
	}
	return reg
}

// Get returns a parser for the file based on its extension, if available.
func (r *ParserRegistry) Get(file string) (pkg.Parser, bool) {
	ext := filepath.Ext(file)
	p, ok := r.parsers[ext]
	return p, ok
}
