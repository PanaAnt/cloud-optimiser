package awsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type MockCloudWatchClient struct{}

// IsMock returns true indicating  mock client
func (m *MockCloudWatchClient) IsMock() bool {
	return true
}

// GetCpuUtilisation reads mock CPU metrics from testdata/metrics.json
func (m *MockCloudWatchClient) GetCpuUtilisation(ctx context.Context, instanceID string, hours int) (model.CPUSampleSeries, error) {
	path := filepath.Join("testdata", "metrics.json")

	file, err := os.ReadFile(path)
	if err != nil {
		return model.CPUSampleSeries{}, fmt.Errorf("failed to read mock metrics: %w", err)
	}

	var data map[string][]float64
	if err := json.Unmarshal(file, &data); err != nil {
		return model.CPUSampleSeries{}, fmt.Errorf("failed to unmarshal mock metrics: %w", err)
	}

	samples, ok := data[instanceID]
	if !ok {
		// Return empty samples if instance not found (not an error)
		samples = []float64{}
	}

	return model.CPUSampleSeries{
		InstanceID: instanceID,
		Samples:    samples,
	}, nil
}