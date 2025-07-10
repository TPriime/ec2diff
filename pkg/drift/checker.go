package drift

import (
	"context"
	"sync"

	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/logger"
)

// driftChecker implements the DriftChecker interface.
type driftChecker struct {
	// number of concurrent workers
	workers int
}

// NewDriftChecker returns a new instance of driftChecker.
func NewDriftChecker(workers int) pkg.DriftChecker {
	return &driftChecker{workers: workers}
}

// CheckDrift compares liveInstances with stateInstances to detect drift.
//
// It returns a list of reports indicating changed or missing attributes.
func (d driftChecker) CheckDrift(ctx context.Context, liveInstances, stateInstances pkg.InstanceMap, attributes []string) []pkg.Report {
	ctx = logger.With(ctx, "op", "drift.CheckDrift")

	jobs := make(chan string, len(liveInstances))
	results := make(chan pkg.Report, len(liveInstances))

	var wg sync.WaitGroup

	// Start a fixed pool of workers
	for i := 0; i < d.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for instanceID := range jobs {
				var report pkg.Report
				logger.Info(ctx, "Comparing live and state for instance", "worker", workerID, "instanceID", instanceID)

				if stateInst, found := stateInstances[instanceID]; found {
					report = compareState(instanceID, instanceToState(liveInstances[instanceID]), instanceToState(stateInst), attributes)
				} else {
					logger.Info(ctx, "Instance missing in state", "worker", workerID, "instanceID", instanceID)
					report = reportMissing(instanceID, instanceToState(liveInstances[instanceID]), attributes)
				}

				results <- report
			}
		}(i)
	}

	// Feed jobs to the queue
	go func() {
		for instanceID := range liveInstances {
			jobs <- instanceID
		}
		close(jobs)
	}()

	// Wait for workers to finish and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect the results
	var reports []pkg.Report
	for r := range results {
		reports = append(reports, r)
	}

	logger.Info(ctx, "Drift reports collected", "reports", len(reports))
	return reports
}
