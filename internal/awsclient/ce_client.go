package awsclient

import (
	"context"
	"fmt"

	"github.com/PanaAnt/cloud-optimiser/internal/config"
	"github.com/PanaAnt/cloud-optimiser/internal/logging"
	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type CostExplorerClient interface {
	GetInstanceCost(ctx context.Context, instanceID string, days int) (model.CostData, error)
	IsMock() bool
}

// NewCostExplorer creates a Cost Explorer client based on the provided configuration.
// Returns an error if real AWS client creation fails and mock mode is not enabled.
func NewCostExplorer(ctx context.Context, cfg Config) (CostExplorerClient, error) {
	// Forced mock via flag
	if cfg.UseMock {
		logging.Debug("Cost Explorer: Using MOCK (flag override)")
		return &MockCostExplorer{}, nil
	}

	appCfg, err := config.LoadConfig()
	if err != nil {
		logging.Warn(fmt.Sprintf("Could not load config: %v, defaulting to mock", err))
		return &MockCostExplorer{}, nil
	}

	if appCfg.Mode == "mock" {
		logging.Debug("Cost Explorer: Using MOCK (from config)")
		return &MockCostExplorer{}, nil
	}

	// Attempt to create real AWS client
	logging.Debug("Cost Explorer: Attempting to create real AWS client")
	client, err := NewRealCostExplorer(ctx, cfg.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create real Cost Explorer client: %w", err)
	}

	logging.Debug("Cost Explorer: Using REAL AWS")
	return client, nil
}
