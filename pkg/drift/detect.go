package drift

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/tpriime/ec2diff/pkg/aws"
)

// AttributeDrift describes an attribute mismatch
type AttributeDrift struct {
	Name     string      `json:"name"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
}

// Report captures drift for one instance
type Report struct {
	InstanceID string           `json:"instance_id"`
	Drifts     []AttributeDrift `json:"drifts"`
}

// CheckDrift compares AWS vs Terraform-state for one instance
func CheckDrift(ctx context.Context,
	client *aws.Client,
	stateMap map[string]map[string]interface{},
	instanceID string,
	attrs []string,
) Report {

	tfAttrs, ok := stateMap[instanceID]
	if !ok {
		return Report{InstanceID: instanceID}
	}

	inst, err := client.GetInstance(ctx, instanceID)
	if err != nil {
		return Report{InstanceID: instanceID}
	}

	drifts := []AttributeDrift{}
	for _, a := range attrs {
		switch strings.ToLower(a) {
		case "instance_type":
			exp := tfAttrs["instance_type"]
			act := string(inst.InstanceType)
			if exp != act {
				drifts = append(drifts, AttributeDrift{
					Name:     "instance_type",
					Expected: exp,
					Actual:   act,
				})
			}

		case "tags", "name":
			// flatten AWS tags
			awsMap := map[string]string{}
			for _, t := range inst.Tags {
				awsMap[*t.Key] = *t.Value
			}
			// Terraform tags come in a map[string]interface{}
			tfRaw, _ := tfAttrs["tags"].(map[string]interface{})
			tfMap := map[string]string{}
			for k, v := range tfRaw {
				tfMap[k] = v.(string)
			}
			if !reflect.DeepEqual(awsMap, tfMap) {
				drifts = append(drifts, AttributeDrift{
					Name:     "tags",
					Expected: tfMap,
					Actual:   awsMap,
				})
			}

		case "sg", "security_groups":
			awsSg := []string{}
			for _, sg := range inst.SecurityGroups {
				awsSg = append(awsSg, *sg.GroupId)
			}
			sort.Strings(awsSg)

			raw, _ := tfAttrs["vpc_security_group_ids"].([]interface{})
			tfSg := []string{}
			for _, v := range raw {
				tfSg = append(tfSg, v.(string))
			}
			sort.Strings(tfSg)

			if !reflect.DeepEqual(awsSg, tfSg) {
				drifts = append(drifts, AttributeDrift{
					Name:     "vpc_security_group_ids",
					Expected: tfSg,
					Actual:   awsSg,
				})
			}

		default:
			// skip unknown attrs
		}
	}

	return Report{
		InstanceID: instanceID,
		Drifts:     drifts,
	}
}
