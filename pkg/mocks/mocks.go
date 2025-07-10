package mocks

import (
	"context"

	"github.com/tpriime/ec2diff/pkg"
)

// MockParser implements pkg.Parser for testing
type MockParser struct {
	Parsed     pkg.InstanceMap
	Extensions []string
	Err        error
}

func (m *MockParser) Parse(path string, ids []string) (pkg.InstanceMap, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Parsed, nil
}

func (m *MockParser) SupportedTypes() []string {
	return m.Extensions
}

// MockLiveFetcher implements pkg.LiveFetcher for testing
type MockLiveFetcher struct {
	Instances pkg.InstanceMap
	Err       error
}

func (m *MockLiveFetcher) Fetch(_ context.Context, onPpageFn func(page int, instances pkg.InstanceMap) bool) error {
	if m.Err != nil {
		return m.Err
	}
	onPpageFn(1, m.Instances)
	return nil
}

// MockReportPrinter implements pkg.ReportPrinter for testing
type MockReportPrinter struct {
	Output []pkg.Report
}

func (m *MockReportPrinter) Print(reports []pkg.Report) {
	m.Output = reports
}

// MockDriftChecker implements pkg.DriftChecker for testing
type MockDriftChecker struct{}

func (m *MockDriftChecker) CheckDrift(ctx context.Context, live, state pkg.InstanceMap, attrs []string) []pkg.Report {
	return []pkg.Report{{InstanceID: "i-abc", Drifts: nil}}
}
