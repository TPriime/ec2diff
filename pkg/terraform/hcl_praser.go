package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/tpriime/ec2diff/pkg"
)

type config struct {
	Resource []awsInstanceBlock `hcl:"resource,block"`
}

type awsInstanceBlock struct {
	Type string `hcl:"type,label"` // e.g., "aws_instance"
	Name string `hcl:"name,label"` // e.g., "example"`

	Ami                 string            `hcl:"ami,optional"`
	InstanceType        string            `hcl:"instance_type"`
	VpcSecurityGroupIds []string          `hcl:"vpc_security_group_ids,optional"`
	KeyName             string            `hcl:"key_name,optional"`
	SecurityGroups      []string          `hcl:"security_groups,optional"`
	Tags                map[string]string `hcl:"tags,optional"`
	InstanceState       string            `hcl:"instance_state,optional"`
	PublicIP            string            `hcl:"public_ip,optional"`
}

func (i awsInstanceBlock) toInstance(id string) pkg.Instance {
	return pkg.Instance{
		ID:             id,
		Type:           i.InstanceType,
		State:          i.InstanceState,
		KeyName:        i.KeyName,
		Tags:           i.Tags,
		SecurityGroups: i.SecurityGroups,
		PublicIP:       i.PublicIP,
	}
}

func ParseHCL(path string, ids []string) (map[string]pkg.Instance, error) {
	var config config

	if len(ids) == 0 {
		return nil, fmt.Errorf("no instance ids provided")
	}

	err := hclsimple.DecodeFile(path, nil, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL file: %v", err)
	}

	if len(ids) > len(config.Resource) {
		return nil, fmt.Errorf("given instance ids exceed found resources")
	}

	out := map[string]pkg.Instance{}
	for i, id := range ids {
		out[id] = config.Resource[i].toInstance(id)
	}

	return out, nil
}
