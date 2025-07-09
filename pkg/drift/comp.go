package drift

import (
	"slices"

	"github.com/google/go-cmp/cmp"
	"github.com/tpriime/ec2diff/pkg"
)

type state map[string]any

// compareState checks two instance states and identifies attribute-level drift.
// It returns a report listing any changed, missing, or unexpected attributes.
func compareState(id string, stateA, stateB state, attrs []string) pkg.Report {
	drifts := []pkg.AttributeDrift{}

	// Compare attributes present in stateA
	for attr, valueA := range stateA {
		// filter to include only selected attributes
		if !slices.Contains(attrs, attr) {
			continue
		}
		valueB, ok := stateB[attr]
		if !ok || !cmp.Equal(valueA, valueB) {
			drifts = append(drifts, pkg.AttributeDrift{Name: attr, Expected: valueA, Found: valueB})
		}
	}

	// Find extra attributes present in stateB but missing in stateA
	for attr := range stateB {
		// filter to include only selected attributes
		if !slices.Contains(attrs, attr) {
			continue
		}
		if _, ok := stateA[attr]; !ok {
			drifts = append(drifts, pkg.AttributeDrift{Name: attr, Expected: "-", Found: stateB[attr]})
		}
	}

	comment := pkg.CommentDriftDetected
	if len(drifts) == 0 {
		comment = pkg.CommentNoDriftDetected
	}

	return pkg.Report{
		InstanceID: id,
		Drifts:     drifts,
		Comment:    comment,
	}
}

// reportMissing generates a drift report for an instance missing from stateB.
// It assumes the instance is present only in stateA and marks all attributes as missing.
func reportMissing(id string, stateA state, attrs []string) pkg.Report {
	drifts := []pkg.AttributeDrift{}
	for attr, value := range stateA {
		// filter to include only selected attributes
		if !slices.Contains(attrs, attr) {
			continue
		}
		drifts = append(drifts, pkg.AttributeDrift{Name: attr, Expected: value, Found: "-"})
	}
	return pkg.Report{
		InstanceID: id,
		Drifts:     drifts,
		Comment:    pkg.CommentMissingState,
	}
}

func instanceToState(i pkg.Instance) state {
	return state{
		pkg.AttrInstanceType:   i.Type,
		pkg.AttrInstanceState:  i.State,
		pkg.AttrKeyName:        i.KeyName,
		pkg.AttrTags:           i.Tags,
		pkg.AttrSecurityGroups: i.SecurityGroups,
		pkg.AttrPublicIP:       i.PublicIP,
	}
}
