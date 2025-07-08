package drift

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
)

func TestCompareState(t *testing.T) {
	a := pkg.Instance{
		ID:   "222",
		Type: "t2.small",
	}
	b := pkg.Instance{
		ID:   "222",
		Type: "t2.micro",
	}
	attrs := []string{"instance_type", "tags", "sg"}

	report := compareState(a.ID, instanceToState(a), instanceToState(b), attrs)
	assert.Len(t, report.Drifts, 1, "expected 1 drift (instance_type)")
}

func TestCompareState_NoDrift(t *testing.T) {
	a := pkg.Instance{
		ID:   "222",
		Type: "t2.small",
	}
	b := pkg.Instance{
		ID:   "222",
		Type: "t2.small",
	}
	attrs := []string{"instance_type", "tags", "sg"}

	report := compareState(a.ID, instanceToState(a), instanceToState(b), attrs)
	assert.Len(t, report.Drifts, 0, "expected 1 drift (instance_type)")
}

func TestCompareState_Attributes(t *testing.T) {
	a := pkg.Instance{
		ID:    "i-222",
		Type:  "t2.micro",
		State: "stopped",
	}
	b := pkg.Instance{
		ID:    "i-222",
		Type:  "t2.small",
		State: "running",
	}

	for name, attrs := range map[string][]string{
		"subset 1": {pkg.AttrInstanceType},
		"subset 2": {pkg.AttrInstanceType, pkg.AttrInstanceState},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			report := compareState(a.ID, instanceToState(a), instanceToState(b), attrs)
			assert.Len(t, report.Drifts, len(attrs), fmt.Sprintf("expected %d drifts", len(attrs)))
			for _, d := range report.Drifts {
				if !slices.Contains(attrs, d.Name) {
					assert.Fail(t, fmt.Sprintf("drift name outside given attributes %v", attrs))
				}
			}
		})
	}
}

func TestCompareState_OnlySelectedAttributes(t *testing.T) {
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

		report := compareState(a.ID, instanceToState(a), instanceToState(b), []string{})
		assert.Len(t, report.Drifts, 0, "should return no drifts")
	})
}

func TestCompareMissing(t *testing.T) {
	a := pkg.Instance{
		ID:      "222",
		Type:    "t2.small",
		State:   "running",
		KeyName: "key_name",
	}

	report := reportMissing(a.ID, instanceToState(a),
		[]string{pkg.AttrInstanceType, pkg.AttrInstanceState, pkg.AttrKeyName})
	assert.Len(t, report.Drifts, 3, "expected 3 drifts")
	for _, d := range report.Drifts {
		assert.Equal(t, "-", d.Found)
	}
}
