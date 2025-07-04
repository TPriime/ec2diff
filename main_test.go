package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
)

var mockInstance = pkg.Instance{
	ID:   "i-123",
	Type: "t2.samll",
}

type mockClient struct{ result pkg.Instance }

func (m mockClient) GetInstance(ctx context.Context, instanceID string) (*pkg.Instance, error) {
	return &m.result, nil
}

func getMockClient() mockClient {
	return mockClient{result: mockInstance}
}

func tempFile(t *testing.T, name string, content string) string {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	assert.NoError(t, err)
	return path
}

func Test_CLI_Args_StateFile(t *testing.T) {
	hclContent := `
	{
		"resources":[
		{
			"type":"aws_instance",
			"instances":[
				{
					"attributes":{
					"id":"i-123",
					"instance_type":"t2.micro"
					}
				}
			]
		}
		]
	}`

	hclPath := tempFile(t, "test.hcl", hclContent)
	cmd := setupCommand(options{
		client: getMockClient(),
	})

	cmd.SetArgs([]string{"--state", hclPath, "--instances", "i-123", "--attrs", "instance_type"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "instance_type")
}

func Test_CLI_Args_HCLFile(t *testing.T) {
	hclContent := `
		resource "aws_instance" "example" {
			ami                  = "ami-abc123"
			instance_type        = "t2.micro"
		}
	`

	hclPath := tempFile(t, "test.hcl", hclContent)
	cmd := setupCommand(options{
		client: getMockClient(),
	})

	cmd.SetArgs([]string{"--hcl", hclPath, "--instances", "i-123", "--attrs", "instance_type"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "instance_type")
}

func Test_CLI_Args_MissingRequired(t *testing.T) {
	cmd := setupCommand(options{
		client: getMockClient(),
	})
	cmd.SetArgs([]string{"--instances", "i-1"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "required")
}
func Test_getTargetInstances_AllInstances(t *testing.T) {
	state := map[string]pkg.Instance{
		"i-1": {ID: "i-1", Type: "t2.micro"},
		"i-2": {ID: "i-2", Type: "t2.small"},
	}
	instances, err := getTargetInstances(state, nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 2)
	ids := []string{instances[0].ID, instances[1].ID}
	assert.Contains(t, ids, "i-1")
	assert.Contains(t, ids, "i-2")
}

func Test_getTargetInstances_SpecificInstances(t *testing.T) {
	state := map[string]pkg.Instance{
		"i-1": {ID: "i-1", Type: "t2.micro"},
		"i-2": {ID: "i-2", Type: "t2.small"},
	}
	instances, err := getTargetInstances(state, []string{"i-2"})

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "i-2", instances[0].ID)
}

func Test_getTargetInstances_InstanceNotFound(t *testing.T) {
	state := map[string]pkg.Instance{
		"i-1": {ID: "i-1", Type: "t2.micro"},
	}
	instances, err := getTargetInstances(state, []string{"i-not-found"})

	assert.Nil(t, instances)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance not found")
}

func Test_getTargetInstances_EmptyState(t *testing.T) {
	state := map[string]pkg.Instance{}
	instances, err := getTargetInstances(state, nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 0)
}

func Test_getTargetInstances_MultipleIDs(t *testing.T) {
	state := map[string]pkg.Instance{
		"i-1": {ID: "i-1", Type: "t2.micro"},
		"i-2": {ID: "i-2", Type: "t2.small"},
		"i-3": {ID: "i-3", Type: "t2.nano"},
	}
	instances, err := getTargetInstances(state, []string{"i-1", "i-3"})

	assert.NoError(t, err)
	assert.Len(t, instances, 2)
	ids := []string{instances[0].ID, instances[1].ID}
	assert.Contains(t, ids, "i-1")
	assert.Contains(t, ids, "i-3")
}
