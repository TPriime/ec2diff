package terraform

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tpriime/ec2diff/pkg"
)

// ParseState reads the JSON state and returns a map from instanceID to attributes
func ParseState(path string) (map[string]pkg.Instance, error) {
	data, err := os.ReadFile(path)
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
