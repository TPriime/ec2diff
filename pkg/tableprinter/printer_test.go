package tableprinter

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
)

func TestReport_Print_NoDrifts(t *testing.T) {
	var buf bytes.Buffer
	printer := tablePrinter{out: &buf}
	report := pkg.Report{
		InstanceID: "i-123456",
		Drifts:     nil,
		Comment:    "No drifts detected for i-123456",
	}

	printer.Print([]pkg.Report{report})

	output := buf.String()
	assert.Contains(t, output, "No drifts detected for i-123456", "expected output to mention no drifts")
}

func TestReport_Print_WithDrifts(t *testing.T) {
	var buf bytes.Buffer
	printer := tablePrinter{out: &buf}
	report := pkg.Report{
		InstanceID: "i-7890",
		Drifts: []pkg.AttributeDrift{
			{
				Name:     "instance_type",
				Expected: "t2.micro",
				Found:    "t2.small",
			},
			{
				Name:     "tags",
				Expected: map[string]string{"env": "prod"},
				Found:    map[string]string{"env": "dev"},
			},
		},
	}

	printer.Print([]pkg.Report{report})

	output := buf.String()
	assert.Contains(t, output, "i-7890", "expected instance id")
	assert.Contains(t, output, "instance_type", "expected instance_type drift row")
	assert.Contains(t, output, "t2.micro", "expected instance_type expected value")
	assert.Contains(t, output, "t2.small", "expected instance_type actual value")
	assert.Contains(t, output, "tags", "expected tags drift row")
	assert.Contains(t, output, `{"env":"prod"}`, "expected tags drift row with JSON (expected)")
	assert.Contains(t, output, `{"env":"dev"}`, "expected tags drift row with JSON (actual)")
}
