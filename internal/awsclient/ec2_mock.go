package awsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type MockClient struct{}

// IsMock returns true indicating mock client
func (m *MockClient) IsMock() bool {
	return true
}

// ListInstances reads mock EC2 instances from testdata/instance_exmpl.json
func (m *MockClient) ListInstances(ctx context.Context) ([]model.EC2Instance, error) {
	path := filepath.Join("testdata", "instance_exmpl.json")

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mock data: %w", err)
	}

	var instances []model.EC2Instance
	if err := json.Unmarshal(file, &instances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mock data: %w", err)
	}

	return instances, nil
}
