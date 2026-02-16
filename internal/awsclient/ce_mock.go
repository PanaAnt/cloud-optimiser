package awsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type MockCostExplorer struct{}

// IsMock returns true indicating mock client
func (m *MockCostExplorer) IsMock() bool {
	return true
}

// GetInstanceCost reads mock cost data from testdata/costs.json
func (m *MockCostExplorer) GetInstanceCost(ctx context.Context, instanceID string, days int) (model.CostData, error) {
	path := filepath.Join("testdata", "costs.json")

	file, err := os.ReadFile(path)
	if err != nil {
		return model.CostData{}, fmt.Errorf("failed to read mock cost data: %w", err)
	}

	var data map[string]model.CostData
	if err := json.Unmarshal(file, &data); err != nil {
		return model.CostData{}, fmt.Errorf("failed to unmarshal mock cost data: %w", err)
	}

	if cost, ok := data[instanceID]; ok {
		return cost, nil
	}

	// Return zero cost if instance not found (not an error)
	return model.CostData{
		InstanceID:  instanceID,
		MonthlyCost: 0,
		HourlyCost:  0,
	}, nil
}
