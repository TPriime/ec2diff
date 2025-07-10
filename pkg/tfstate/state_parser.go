// tfStateParser is a parser for Terraform .tfstate files, extracting aws_instance resources.
package tfstate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tpriime/ec2diff/pkg"
)

// tfStateParser implements the Parser interface for .tfstate files.
type tfStateParser struct{}

// NewTfStateParser creates a new tfStateParser instance.
func NewTfStateParser() pkg.Parser {
	return &tfStateParser{}
}

// Parse loads a Terraform state file and maps matching aws_instance resources by ID.
func (t tfStateParser) Parse(filePath string) (map[string]pkg.Instance, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}

	out := map[string]pkg.Instance{}
	for _, res := range st.Resources {
		if res.Type != "aws_instance" {
			continue
		}
		for _, inst := range res.Instances {
			out[inst.Attributes.ID] = inst.toInstance()
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no aws_instance resources found in state")
	}
	return out, nil
}

// SupportedTypes returns the file extensions this parser handles.
func (tfStateParser) SupportedTypes() []string {
	return []string{".tfstate", ".json"}
}
