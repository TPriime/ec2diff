package drift

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
)

// mockState creates a simple instance used in drift tests.
func mockState(id, instType, state, key string) pkg.Instance {
	return pkg.Instance{
		ID:             id,
		Type:           instType,
		State:          state,
		KeyName:        key,
		Tags:           map[string]string{"env": "test"},
		SecurityGroups: []string{"sg-1"},
		PublicIP:       "1.2.3.4",
	}
}

func TestCheckDrift_NoDrift(t *testing.T) {
	live := pkg.InstanceMap{"i-1": mockState("i-1", "t2.micro", "running", "my-key")}
	state := pkg.InstanceMap{"i-1": mockState("i-1", "t2.micro", "running", "my-key")}

	reports := NewDriftChecker(2).CheckDrift(t.Context(), live, state, []string{"Type", "State", "KeyName"})

	assert.Len(t, reports, 1)
	assert.Empty(t, reports[0].Drifts)
}

func TestCheckDrift_AttributeDrift(t *testing.T) {
	live := pkg.InstanceMap{"i-1": mockState("i-1", "t2.micro", "stopped", "my-key")}
	state := pkg.InstanceMap{"i-1": mockState("i-1", "t2.micro", "running", "my-key")}

	reports := NewDriftChecker(2).CheckDrift(t.Context(), live, state, []string{pkg.AttrInstanceState})

	assert.Len(t, reports, 1)
	assert.Len(t, reports[0].Drifts, 1)
	assert.Equal(t, pkg.AttrInstanceState, reports[0].Drifts[0].Name)
}

func TestCheckDrift_MissingInstance(t *testing.T) {
	live := pkg.InstanceMap{
		"i-1": mockState("i-1", "t2.micro", "running", "key"),
		"i-2": mockState("i-2", "t3.small", "stopped", "key"),
	}
	state := pkg.InstanceMap{
		"i-1": mockState("i-1", "t2.micro", "running", "key"),
	}

	reports := NewDriftChecker(2).CheckDrift(t.Context(), live, state,
		[]string{pkg.AttrInstanceType, pkg.AttrInstanceState})

	assert.Len(t, reports, 2)

	found := false
	for _, r := range reports {
		if r.InstanceID == "i-2" && r.Comment == pkg.CommentMissingState {
			found = true
		}
	}
	assert.True(t, found, "expected report for missing instance i-2")
}

func TestCheckDrift_CustomAttributes(t *testing.T) {
	live := pkg.InstanceMap{"i-1": mockState("i-1", "t2.micro", "running", "key1")}
	state := pkg.InstanceMap{"i-1": mockState("i-1", "t2.micro", "running", "key2")}

	reports := NewDriftChecker(2).CheckDrift(t.Context(), live, state, []string{"State"})

	assert.Empty(t, reports[0].Drifts)
}
