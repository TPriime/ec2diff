package pkg

import "context"

// DriftChecker defines how instances are compared to detect drifts.
type DriftChecker interface {
	CheckDrift(ctx context.Context, liveInstances, targetInstances InstanceMap, attributes []string) []Report
}
