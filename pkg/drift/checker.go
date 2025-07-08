package drift

import (
	"context"
	"sync"

	"github.com/tpriime/ec2diff/pkg"
)

// driftChecker implements the DriftChecker interface.
type driftChecker struct{}

// NewDriftChecker returns a new instance of driftChecker.
func NewDriftChecker() pkg.DriftChecker {
	return &driftChecker{}
}

// CheckDrift compares liveInstances with stateInstances to detect drift.
//
// It returns a list of reports indicating changed or missing attributes.
func (d driftChecker) CheckDrift(ctx context.Context, liveInstances, stateInstances pkg.InstanceMap, attributes []string) []pkg.Report {
	// Channel to collect drift reports safely from goroutines.
	results := make(chan pkg.Report, len(stateInstances))
	var wg sync.WaitGroup

	// Launch drift checks concurrently for each live instance.
	for instanceID := range liveInstances {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			var report pkg.Report

			// Compare against state instance if found, else report as missing.
			if stateInst, found := stateInstances[id]; found {
				report = compareState(id, instanceToState(liveInstances[id]), instanceToState(stateInst), attributes)
			} else {
				report = reportMissing(id, instanceToState(liveInstances[id]), attributes)
			}

			results <- report
		}(instanceID)
	}

	// Close results channel once all goroutines finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Gather and return all reports.
	var reports []pkg.Report
	for r := range results {
		reports = append(reports, r)
	}
	return reports
}
