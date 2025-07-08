package drift

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/tpriime/ec2diff/pkg"
)

// CheckDrift concurrently checks for configuration drift between a list of target EC2 instances
// and their corresponding instances retrieved from AWS using the provided API client. It compares
// the specified attributes for each instance and returns a slice of Report objects summarizing
// the drift status. If an instance is not found on AWS, it is treated as deleted and compared
// against an empty instance. Any errors encountered during retrieval (other than not found)
// will cause the function to panic.
//
// Parameters:
//   - api: An implementation of pkg.Client used to fetch instance data from AWS.
//   - targetInstances: A slice of pkg.Instance representing the desired state of instances.
//   - attributes: A slice of strings specifying which attributes to compare for drift.
//
// Returns:
//   - A slice of Report objects, each representing the drift comparison result for an instance.
func CheckDrift(api pkg.LiveFetcher, targetInstances []pkg.Instance, attributes []string) []Report {
	// drift detection concurrently
	var wg sync.WaitGroup
	results := make(chan Report, len(targetInstances))
	errChan := make(chan error, len(targetInstances))

	// find instances on aws and check for drift
	ctx := context.Background()
	for _, instance := range targetInstances {
		wg.Add(1)
		go func(target pkg.Instance) {
			defer wg.Done()
			awsIntance, err := api.GetInstance(ctx, target.ID)
			if err != nil {
				if errors.Is(err, pkg.ErrNotFound) {
					awsIntance = &pkg.Instance{} // instance is deleted, use empty object
				} else {
					errChan <- fmt.Errorf("failed to fetch instance %s from remote: %v", target.ID, err)
					return
				}
			}
			report := CompareInstances(ctx, target, *awsIntance, attributes)
			results <- report
		}(instance)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errChan)
	}()

	reports := []Report{}
	// Collect results and errors efficiently
	for i := 0; i < len(targetInstances); i++ {
		select {
		case r := <-results:
			reports = append(reports, r)
		case err := <-errChan:
			panic(err)
		}
	}
	return reports
}
