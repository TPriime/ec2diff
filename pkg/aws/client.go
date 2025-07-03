package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

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
func NewClient(region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	svc := ec2.NewFromConfig(cfg)
	return &Client{EC2: svc}, nil
}

// GetInstance fetches a single EC2 instance by ID
func (c *Client) GetInstance(ctx context.Context, instanceID string) (*types.Instance, error) {
	out, err := c.EC2.DescribeInstances(ctx,
		&ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		})
	if err != nil {
		return nil, err
	}
	for _, res := range out.Reservations {
		for _, inst := range res.Instances {
			if *inst.InstanceId == instanceID {
				return &inst, nil
			}
		}
	}
	return nil, fmt.Errorf("instance %s not found", instanceID)
}
