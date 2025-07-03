package drift

import (
  "context"
  "testing"

  "github.com/aws/aws-sdk-go-v2/service/ec2"
  "github.com/aws/aws-sdk-go-v2/service/ec2/types"
  "github.com/tpriime/ec2diff/pkg/aws"
)

type mockEC2 struct{}
func (m *mockEC2) DescribeInstances(ctx context.Context, in *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
  return &ec2.DescribeInstancesOutput{
    Reservations: []types.Reservation{
      {
        Instances: []types.Instance{
          {
            InstanceId:   ptrStr("i-1"),
            InstanceType: types.InstanceTypeT2Small,
            Tags: []types.Tag{{Key: ptrStr("Name"), Value: ptrStr("foo")}},
            SecurityGroups: []types.GroupIdentifier{{GroupId: ptrStr("sg-1")}},
          },
        },
      },
    },
  }, nil
}

func ptrStr(s string) *string {
	return &s
}

func TestCheckDrift(t *testing.T) {
  stateMap := map[string]map[string]interface{}{
    "i-1": {
      "instance_type":            "t2.micro",
      "tags":                     map[string]interface{}{"Name": "foo"},
      "vpc_security_group_ids":   []interface{}{"sg-1"},
    },
  }
  client := &aws.Client{EC2: &mockEC2{}}
  report := CheckDrift(context.Background(), client, stateMap, "i-1", []string{"instance_type", "tags", "sg"})
  if len(report.Drifts) != 1 {
    t.Errorf("expected 1 drift (instance_type), got %d", len(report.Drifts))
  }
}
