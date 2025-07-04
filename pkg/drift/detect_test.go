package drift

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
)

// MockClient implements pkg.Client for testing
type MockClient struct {
	Instances map[string]*pkg.Instance
	Err       error
	mu        sync.Mutex
	Calls     []string
}

func (m *MockClient) GetInstance(ctx context.Context, id string) (*pkg.Instance, error) {
	m.mu.Lock()
	m.Calls = append(m.Calls, id)
	m.mu.Unlock()
	if m.Err != nil {
		return nil, m.Err
	}
	inst, ok := m.Instances[id]
	if !ok {
		return nil, pkg.ErrNotFound
	}
	return inst, nil
}

func TestCheckDrift_AllInstancesFound_ReturnsReports(t *testing.T) {
	client := &MockClient{
		Instances: map[string]*pkg.Instance{
			"i-1": {ID: "i-1", Type: "t3.micro"},
			"i-2": {ID: "i-2"},
		},
	}
	targets := []pkg.Instance{
		{ID: "i-1", Type: "t2.small"},
		{ID: "i-2", State: "running"},
	}

	reports := CheckDrift(client, targets, []string{})
	
	assert.Len(t, reports, 2)
	for _, r := range reports {
		assert.Contains(t, []string{"i-1", "i-2"}, r.InstanceID)
		assert.NotEmpty(t, r.Drifts, r.InstanceID+" should have drifts")
	}
}

func TestCheckDrift_InstanceNotFound_ReturnsDeletedReport(t *testing.T) {
	client := &MockClient{
		Instances: map[string]*pkg.Instance{},
	}
	targets := []pkg.Instance{
		{ID: "i-missing", Type: "t2.small"},
	}
	
	reports := CheckDrift(client, targets, []string{pkg.InstanceType})
	
	assert.Len(t, reports, 1)
	assert.Equal(t, "t2.small", reports[0].Drifts[0].Expected)
	assert.Equal(t, "", reports[0].Drifts[0].Actual)
}

func TestCheckDrift_RemoteError_Fatals(t *testing.T) {
	client := &MockClient{
		Err: errors.New("unexpected"),
	}
	targets := []pkg.Instance{
		{ID: "i-err"},
	}
	
	assert.Panics(t, func() {
		CheckDrift(client, targets, []string{"Type"})
	}, "expected fatal error")
}

func TestCheckDrift_HandleMissingInstance(t *testing.T) {
	client := &MockClient{
		Err: pkg.ErrNotFound,
	}
	targets := []pkg.Instance{
		{ID: "i-deleted", Type: "t2.small"},
	}

	reports := CheckDrift(client, targets, []string{pkg.InstanceType})
	
	assert.Len(t, reports, 1)
	assert.Equal(t, "t2.small", reports[0].Drifts[0].Expected)
	assert.Equal(t, "", reports[0].Drifts[0].Actual)
}
