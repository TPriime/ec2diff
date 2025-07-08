package aws

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/stretchr/testify/assert"
)

type MockEC2API struct {
	mockReponseJson string
}

func (m *MockEC2API) DescribeInstances(ctx context.Context, _ *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	data := []byte(m.mockReponseJson)

	var output ec2.DescribeInstancesOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func NewMockClient(responseJson string) *awsFetcher {
	return &awsFetcher{client: &MockEC2API{mockReponseJson: responseJson}}
}

func TestGetInstance_Success(t *testing.T) {
	client := NewMockClient(`
{
  "Reservations": [
    {
      "Instances": [
        {
          "InstanceId": "i-0123456789abcdef0",
          "InstanceType": "t2.small",
		  "KeyName": "test-key",
          "Tags": [
            { "Key": "Name", "Value": "test-instance" },
            { "Key": "Env",  "Value": "staging" }
          ],
          "SecurityGroups": [
            { "GroupId": "sg-0123456789abcdef0", "GroupName": "default" },
            { "GroupId": "sg-0fedcba9876543210", "GroupName": "extra-sg" }
          ]
        }
      ]
    }
  ]
}`)
	result, err := client.Fetch(t.Context())

	assert.NoError(t, err)
	assert.Contains(t, result, "i-0123456789abcdef0")
	inst := result["i-0123456789abcdef0"]
	assert.Equal(t, "t2.small", inst.Type)
	assert.Equal(t, "test-key", inst.KeyName)
	assert.Equal(t, "test-instance", inst.Tags["Name"])
	assert.Contains(t, inst.SecurityGroups, "default")
	assert.Contains(t, inst.SecurityGroups, "extra-sg")
}

func TestGetInstance_EmptyResponse(t *testing.T) {
	client := NewMockClient(`
{
  "Reservations": [
    {
      "Instances": []
    }
  ]
}`,
	)

	result, err := client.Fetch(t.Context())

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}
