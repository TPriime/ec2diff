package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAwsInstanceBlockToInstance(t *testing.T) {
	block := awsInstanceBlock{
		Type:                "aws_instance",
		Name:                "example",
		Ami:                 "ami-abc123",
		InstanceType:        "t2.micro",
		VpcSecurityGroupIds: []string{"sg-1", "sg-2"},
		KeyName:             "my-key",
		SecurityGroups:      []string{"sg-3"},
		Tags:                map[string]string{"Name": "test"},
		InstanceState:       "running",
		PublicIP:            "1.2.3.4",
	}
	id := "i-123"
	inst := block.toInstance(id)
	assert.Equal(t, id, inst.ID)
	assert.Equal(t, "t2.micro", inst.Type)
	assert.Equal(t, "running", inst.State)
	assert.Equal(t, "my-key", inst.KeyName)
	assert.Equal(t, map[string]string{"Name": "test"}, inst.Tags)
	assert.Equal(t, []string{"sg-3"}, inst.SecurityGroups)
	assert.Equal(t, "1.2.3.4", inst.PublicIP)
}

func TestParseHCL_Success(t *testing.T) {
	hclContent := `
resource "aws_instance" "example" {
  ami                  = "ami-abc123"
  instance_type        = "t2.micro"
  vpc_security_group_ids = ["sg-1", "sg-2"]
  key_name             = "my-key"
  security_groups      = ["sg-3"]
  tags = {
	Name = "test"
  }
  instance_state = "running"
  public_ip      = "1.2.3.4"
}
`
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, "main.hcl")
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	assert.NoError(t, err)

	ids := []string{"i-123"}
	instances, err := ParseHCL(hclPath, ids)
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	inst, ok := instances["i-123"]
	assert.True(t, ok)
	assert.Equal(t, "t2.micro", inst.Type)
	assert.Equal(t, "running", inst.State)
	assert.Equal(t, "my-key", inst.KeyName)
	assert.Equal(t, map[string]string{"Name": "test"}, inst.Tags)
	assert.Equal(t, []string{"sg-3"}, inst.SecurityGroups)
	assert.Equal(t, "1.2.3.4", inst.PublicIP)
}

func TestParseHCL_NoIDs(t *testing.T) {
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, "main.hcl")
	os.WriteFile(hclPath, []byte(""), 0644)
	_, err := ParseHCL(hclPath, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no instance ids provided")
}

func TestParseHCL_TooManyIDs(t *testing.T) {
	hclContent := `
resource "aws_instance" "example" {
  ami = "ami-abc123"
  instance_type = "t2.micro"
}
`
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, "main.hcl")
	os.WriteFile(hclPath, []byte(hclContent), 0644)
	_, err := ParseHCL(hclPath, []string{"i-1", "i-2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "given instance ids exceed found resources")
}

func TestParseHCL_InvalidFile(t *testing.T) {
	_, err := ParseHCL("nonexistent.tf", []string{"i-1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse HCL file")
}
