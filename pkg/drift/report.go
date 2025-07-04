package drift

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

// Report captures drift for one instance
type Report struct {
	InstanceID string           `json:"instance_id"`
	Drifts     []AttributeDrift `json:"drifts"`
}

func (r Report) Print(output io.Writer) {
	w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)

	if len(r.Drifts) == 0 {
		fmt.Fprintf(w, "No drifts detected for %s", r.InstanceID)
		w.Flush()
		return
	}

	fmt.Fprintf(w, "Instance: %s\n", r.InstanceID)
	fmt.Fprintln(w, "Attribute\tExpected\tAWS")
	fmt.Fprintln(w, "---------\t--------\t---")
	for _, d := range r.Drifts {
		if isMapOrSlice(d.Expected) {
			d.Expected = toJSONString(d.Expected)
		}
		if isMapOrSlice(d.Actual) {
			d.Actual = toJSONString(d.Actual)
		}
		fmt.Fprintf(w, "%s\t%v\t%v\n", d.Name, d.Expected, d.Actual)
	}
	fmt.Fprintln(w)
	w.Flush()
}

func isMapOrSlice(v interface{}) bool {
	kind := reflect.TypeOf(v).Kind()
	return kind == reflect.Map || kind == reflect.Slice
}

func toJSONString(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bytes)
}
