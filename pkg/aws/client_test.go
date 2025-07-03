package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

// mockEC2API implements aws.EC2Client
type mockEC2API struct {
	output *ec2.DescribeInstancesOutput
	err    error
}

func (m *mockEC2API) DescribeInstances(_ context.Context, _ *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.output, m.err
}

func TestGetInstanceFromClient_Success(t *testing.T) {
	mockEC2 := &mockEC2API{
		output: &ec2.DescribeInstancesOutput{
			Reservations: []ec2Types.Reservation{
				{
					Instances: []ec2Types.Instance{
						{
							InstanceId:   ptrStr("i-1234567890abcdef0"),
							InstanceType: ec2Types.InstanceTypeT2Micro,
							ImageId:      ptrStr("ami-123"),
							KeyName:      ptrStr("test-key"),
							Monitoring:   &ec2Types.Monitoring{State: "disabled"},
							Tags: []ec2Types.Tag{
								{Key: ptrStr("Name"), Value: ptrStr("test-instance")},
							},
							State: &ec2Types.InstanceState{
								Code: nil,
								Name: "running",
							},
							Architecture:       ec2Types.ArchitectureValuesX8664,
							VirtualizationType: ec2Types.VirtualizationTypeHvm,
							BlockDeviceMappings: []ec2Types.InstanceBlockDeviceMapping{
								{
									DeviceName: ptrStr("/dev/xvda"),
									Ebs:        &ec2Types.EbsInstanceBlockDevice{VolumeId: ptrStr("vol-abc123")},
								},
							},
						},
					},
				},
			},
		},
	}

	client := &Client{EC2: mockEC2}
	result, err := client.GetInstance(context.Background(), "i-1234567890abcdef0")

	assert.NoError(t, err)
	assert.Equal(t, "i-1234567890abcdef0", result.ID)
	assert.Equal(t, "t2.micro", result.Type)
	// assert.Equal(t, "ami-123", result.ImageID)
	assert.Equal(t, "test-key", result.KeyName)
	assert.Equal(t, "test-instance", result.Tags["Name"])
}

func ptrStr(s string) *string {
	return &s
}

func TestGetInstanceFromClient_NotFound(t *testing.T) {
	client := &Client{
		EC2: &mockEC2API{
			output: &ec2.DescribeInstancesOutput{
				Reservations: []ec2Types.Reservation{},
			},
		}}

	_, err := client.GetInstance(context.Background(), "i-missing")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetInstanceFromClient_DescribeError(t *testing.T) {
	client := &Client{
		EC2: &mockEC2API{
			output: nil,
			err:    assert.AnError,
		}}

	_, err := client.GetInstance(context.Background(), "i-err")
	assert.Error(t, err)
}
