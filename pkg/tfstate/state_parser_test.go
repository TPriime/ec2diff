package tfstate

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseState(t *testing.T) {
	parser := &tfStateParser{}

	content := `
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
			},
			{
				"type":"aws_instance",
				"instances":[
					{
						"attributes":{
							"id":"i-125",
							"instance_type":"t2.large"
						}
					}
				]
			}
		]
	}`

	extensions := []string{".tfstate", ".json"}
	for _, ext := range extensions {
		t.Run("should parse extention "+ext, func(t *testing.T) {
			tmp, err := os.CreateTemp("", "state*"+ext)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmp.Name())
			tmp.Write([]byte(content))
			tmp.Close()

			instances, err := parser.Parse(tmp.Name(), nil)

			assert.NoError(t, err)
			assert.Len(t, instances, 2)
			assert.Contains(t, instances, "i-123")
			assert.Contains(t, instances, "i-125")
			assert.Equal(t, "t2.micro", instances["i-123"].Type)
			assert.Equal(t, "t2.large", instances["i-125"].Type)
		})
	}

	t.Run("should filter out ids when supplied", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "state*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmp.Name())
		tmp.Write([]byte(content))
		tmp.Close()

		wanted := "i-125"
		instances, err := parser.Parse(tmp.Name(), []string{wanted})

		assert.NoError(t, err)
		assert.Len(t, instances, 1)
		assert.Contains(t, instances, wanted)
		assert.Equal(t, "t2.large", instances[wanted].Type)
	})
}

func TestSupportedTypes(t *testing.T) {
	sp := &tfStateParser{}
	types := sp.SupportedTypes()

	assert.Len(t, types, 2)
	assert.Contains(t, sp.SupportedTypes(), ".tfstate")
	assert.Contains(t, sp.SupportedTypes(), ".json")
}
