package terraform

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseState(t *testing.T) {
	content := `{
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
	tmp, err := os.CreateTemp("", "state*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Write([]byte(content))
	tmp.Close()

	instances, err := ParseState(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, instances, 1)
	assert.Contains(t, instances, "i-123")
	assert.Equal(t, "t2.micro", instances["i-123"].Type)
}
