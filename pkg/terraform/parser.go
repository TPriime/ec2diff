package terraform

import (
	"encoding/json"
	"fmt"
	"os"
)

// State mirrors the Terraform state JSON structure minimally
type State struct {
	Resources []struct {
		Type      string `json:"type"`
		Instances []struct {
			Attributes map[string]interface{} `json:"attributes"`
		} `json:"instances"`
	} `json:"resources"`
}

// ParseState reads the JSON state and returns a map from instanceID to attributes
func ParseState(path string) (map[string]map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	out := make(map[string]map[string]interface{})
	for _, res := range st.Resources {
		if res.Type != "aws_instance" {
			continue
		}
		for _, inst := range res.Instances {
			idRaw, ok := inst.Attributes["id"]
			if !ok {
				continue
			}
			id, ok := idRaw.(string)
			if !ok {
				continue
			}
			out[id] = inst.Attributes
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no aws_instance resources found in state")
	}
	return out, nil
}
