package terraform

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tpriime/ec2diff/pkg"
)

// state mirrors the Terraform state JSON structure minimally
type state struct {
	Resources []struct {
		Type      string       `json:"type"`
		Instances []tfInstance `json:"instances"`
	} `json:"resources"`
}

type tfInstance struct {
	Attributes struct {
		ID                  string            `json:"id"`
		Ami                 string            `json:"ami"`
		AvailabilityZone    string            `json:"availability_zone"`
		InstanceType        string            `json:"instance_type"`
		InstanceState       string            `json:"instance_state"`
		KeyName             string            `json:"key_name"`
		Monitoring          bool              `json:"monitoring"`
		PublicIP            string            `json:"public_ip"`
		SubnetID            string            `json:"subnet_id"`
		SecurityGroups      []string          `json:"security_groups"`
		VpcID               string            `json:"vpc_id"`
		VpcSecurityGroupIds []string          `json:"vpc_security_group_ids"`
		Tags                map[string]string `json:"tags"`
		Architecture        string            `json:"architecture"`
		VirtualizationType  string            `json:"virtualization_type"`
		IamInstanceProfile  string            `json:"iam_instance_profile"`
	} `json:"attributes"`
}

func (i tfInstance) toInstance() pkg.Instance {
	attr := i.Attributes
	return pkg.Instance{
		ID:             attr.ID,
		Type:           attr.InstanceType,
		State:          attr.InstanceState,
		KeyName:        attr.KeyName,
		Tags:           attr.Tags,
		SecurityGroups: attr.SecurityGroups,
		PublicIP:       attr.PublicIP,
	}
}

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
