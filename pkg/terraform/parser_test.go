package terraform

import (
	"os"
	"testing"
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

	m, err := ParseState(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	if m["i-123"]["instance_type"] != "t2.micro" {
		t.Errorf("expected t2.micro, got %v", m["i-123"]["instance_type"])
	}
}
