package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tpriime/ec2diff/pkg"
)

var ErrNotFound = pkg.ErrNotFound

// EC2API defines methods we need from AWS
type EC2API interface {
	DescribeInstances(
		ctx context.Context,
		params *ec2.DescribeInstancesInput,
		optFns ...func(*ec2.Options),
	) (*ec2.DescribeInstancesOutput, error)
}

// Client wraps the real AWS EC2 client
type Client struct {
	EC2 EC2API
}

// NewClient loads AWS config and returns a Client
func NewClient(region string) (pkg.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	svc := ec2.NewFromConfig(cfg)
	return &Client{EC2: svc}, nil
}

// GetInstance fetches a single EC2 instance by ID
func (c *Client) GetInstance(ctx context.Context, instanceID string) (*pkg.Instance, error) {
	out, err := c.EC2.DescribeInstances(ctx,
		&ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}},
	)
	if err != nil {
		return nil, err
	}
	for _, res := range out.Reservations {
		for _, inst := range res.Instances {
			if *inst.InstanceId == instanceID {
				return typeToInstance(inst), nil
			}
		}
	}
	return nil, ErrNotFound
}

func typeToInstance(inst types.Instance) *pkg.Instance {
	tags := map[string]string{}
	for _, t := range inst.Tags {
		tags[valstr(t.Key)] = valstr(t.Value)
	}

	var state string
	if inst.State != nil {
		state = string(inst.State.Name)
	}

	sgs := []string{}
	for _, sg := range inst.SecurityGroups {
		sgs = append(sgs, valstr(sg.GroupName))
	}

	return &pkg.Instance{
		ID:             valstr(inst.InstanceId),
		Type:           string(inst.InstanceType),
		State:          state,
		KeyName:        valstr(inst.KeyName),
		Tags:           tags,
		SecurityGroups: sgs,
		PublicIP:       valstr(inst.PublicIpAddress),
	}
}

// valstr safely converts pointer strings to string
func valstr(ptr *string) (s string) {
	if ptr != nil {
		s = *ptr
	}
	return
}
