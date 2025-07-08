package tableprinter

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"

	"github.com/tpriime/ec2diff/pkg"
)

// Report captures drift for one instance
type tablePrinter struct {
	out io.Writer
}

func NewTablePrinter(output io.Writer) pkg.ReportPrinter {
	return &tablePrinter{out: output}
}

func (t tablePrinter) Print(reports []pkg.Report) {
	w := tabwriter.NewWriter(t.out, 0, 0, 2, ' ', 0)

	// Print report header with decoration
	fmt.Fprintln(w, "==============================")
	fmt.Fprintln(w, "            REPORT")
	fmt.Fprintf(w, "==============================\n\n")

	for i, r := range reports {
		// Print instance ID and optional comment
		fmt.Fprintf(w, "Instance [%d]\t: %s\n", i+1, r.InstanceID)
		fmt.Fprintf(w, "Comment      \t: %s\n", r.Comment)

		// Print header and drift entries
		if len(r.Drifts) != 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Attribute       \tLive                               \tState")
			fmt.Fprintln(w, "-------------   \t----------------------------------   \t------------------------------")
		}

		for _, d := range r.Drifts {
			expected := d.Expected
			found := d.Found

			if isMapOrSlice(expected) {
				expected = toJSONString(expected)
			}
			if isMapOrSlice(found) {
				found = toJSONString(found)
			}

			fmt.Fprintf(w, "%-15s\t%-35v\t%-30v\n", d.Name, expected, found)
		}

		// Separate instance reports with spacing and em dash line
		if i < len(reports)-1 {
			fmt.Fprintf(w, "\n—\n\n")
		}
	}
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
