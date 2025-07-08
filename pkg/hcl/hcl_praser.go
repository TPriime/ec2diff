package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/tpriime/ec2diff/pkg"
)

// hclParser implements the pkg.Parser interface for decoding HCL/Terraform files
type hclParser struct{}

// NewHclParser returns a new instance of the HCL parser
func NewHclParser() pkg.Parser {
	return &hclParser{}
}

// Parse reads an HCL file at the given path and maps specified instance IDs
// to their corresponding EC2 instance definitions extracted from the resource block.
//
// It expects that the user provides a fixed list of instance IDs, which will
// be used as keys when mapping parsed resources. The order of these IDs should match
// the order of EC2 resources defined in the HCL file.
func (h hclParser) Parse(path string, ids []string) (pkg.InstanceMap, error) {
	var config config // config represents the HCL structure, with a 'Resource' field

	if len(ids) == 0 {
		return nil, fmt.Errorf("no instance ids provided")
	}

	// Decode the HCL file into the config struct
	err := hclsimple.DecodeFile(path, nil, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL file: %v", err)
	}

	// Ensure the caller hasn’t passed more instance IDs than resources available
	if len(ids) > len(config.Resource) {
		return nil, fmt.Errorf("given instance ids exceed found resources")
	}

	out := map[string]pkg.Instance{}

	// Map each instance ID to the corresponding parsed EC2 resource
	for i, id := range ids {
		out[id] = config.Resource[i].toInstance(id)
	}

	return out, nil
}

func (hclParser) SupportedTypes() []string {
	return []string{".hcl", ".tf"} // Supports both HCL and Terraform config files
}
