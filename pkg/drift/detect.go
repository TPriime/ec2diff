package drift

import (
	"context"
	"slices"

	"github.com/google/go-cmp/cmp"
	"github.com/tpriime/ec2diff/pkg"
)

// AttributeDrift describes an attribute mismatch
type AttributeDrift struct {
	Name     string `json:"name"`
	Expected any    `json:"expected"`
	Actual   any    `json:"actual"`
}

// CheckDrift compares AWS vs Terraform-state for one instance
func CheckDrift(ctx context.Context,
	instanceA pkg.Instance,
	instanceB pkg.Instance,
	attrs []string,
) Report {
	if len(attrs) == 0 {
		attrs = pkg.SupportedAttributes()
	}

	drifts := comp(instanceA.ToState(), instanceB.ToState(), attrs)

	return Report{
		InstanceID: instanceA.ID,
		Drifts:     drifts,
	}
}

// Compare maps and return mismatched keys
func comp(a, b pkg.State, attrs []string) []AttributeDrift {
	drifts := []AttributeDrift{}
	for attr, valueA := range a {
		if !slices.Contains(attrs, attr) { // ignore non-specified attributes
			continue
		}
		valueB, ok := b[attr]
		if !ok || !cmp.Equal(valueA, valueB) {
			drifts = append(drifts, AttributeDrift{Name: attr, Expected: a[attr], Actual: b[attr]})
		}
	}

	// check for new attributes in B, missing in A
	for attr := range b {
		if !slices.Contains(attrs, attr) { // ignore non-specified attributes
			continue
		}
		if _, ok := a[attr]; !ok {
			drifts = append(drifts, AttributeDrift{Name: attr, Expected: "<empty>", Actual: b[attr]})
		}
	}
	return drifts
}
