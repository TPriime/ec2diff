package drift

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
)

func TestCheckDrift(t *testing.T) {
	a := pkg.Instance{
		ID:   "222",
		Type: "t2.small",
	}
	b := pkg.Instance{
		ID:   "222",
		Type: "t2.micro",
	}
	
	report := CompareInstances(context.Background(), a, b, []string{"instance_type", "tags", "sg"})
	assert.Len(t, report.Drifts, 1, "expected 1 drift (instance_type)")
}

func TestCheckDrift_NoDrift(t *testing.T) {
	a := pkg.Instance{
		ID:   "222",
		Type: "t2.small",
	}
	b := pkg.Instance{
		ID:   "222",
		Type: "t2.small",
	}

	report := CompareInstances(context.Background(), a, b, []string{"instance_type", "tags", "sg"})
	assert.Len(t, report.Drifts, 0, "expected 1 drift (instance_type)")
}

func TestCheckDrift_Attributes(t *testing.T) {
	a := pkg.Instance{
		ID:    "222",
		Type:  "t2.micro",
		State: "stopped",
	}
	b := pkg.Instance{
		ID:    "222",
		Type:  "t2.small",
		State: "running",
	}

	for name, attrs := range map[string][]string{
		"subset 1": {pkg.InstanceType},
		"subset 2": {pkg.InstanceType, pkg.InstanceState},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			report := CompareInstances(t.Context(), a, b, attrs)
			assert.Len(t, report.Drifts, len(attrs))
			for _, d := range report.Drifts {
				if !slices.Contains(attrs, d.Name) {
					assert.Fail(t, fmt.Sprintf("drift name outside given attributes %v", attrs))
				}
			}
		})
	}

	t.Run("should check all attributes if empty list is given", func(t *testing.T) {
		report := CompareInstances(t.Context(), a, b, []string{})
		assert.Len(t, report.Drifts, 2, "should detect 2 differences")
	})
}

func TestCheckDrift_AttributesAll(t *testing.T) {
	t.Run("should check all attributes if empty list is given", func(t *testing.T) {
		a := pkg.Instance{
			ID:    "222",
			Type:  "t2.micro",
			State: "stopped",
		}
		b := pkg.Instance{
			ID:    "222",
			Type:  "t2.small",
			State: "running",
		}

		report := CompareInstances(t.Context(), a, b, []string{})
		assert.Len(t, report.Drifts, 2, "should detect 2 differences")
	})
}
