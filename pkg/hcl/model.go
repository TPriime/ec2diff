package hcl

import "github.com/tpriime/ec2diff/pkg"

type config struct {
	Resource []awsInstanceBlock `hcl:"resource,block"`
}

type awsInstanceBlock struct {
	Type string `hcl:"type,label"` // e.g., "aws_instance"
	Name string `hcl:"name,label"` // e.g., "example_server"`

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
