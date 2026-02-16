package awsclient

import (
	"context"
	"fmt"

	"github.com/PanaAnt/cloud-optimiser/internal/config"
	"github.com/PanaAnt/cloud-optimiser/internal/logging"
	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type CloudWatchClient interface {
	GetCpuUtilisation(ctx context.Context, instanceID string, hours int) (model.CPUSampleSeries, error)
	IsMock() bool
}

// NewCloudWatch creates a CloudWatch client based on the provided configuration.
// Returns an error if real AWS client creation fails and mock mode is not enabled.
func NewCloudWatch(ctx context.Context, cfg Config) (CloudWatchClient, error) {
	// Forced mock via flag
	if cfg.UseMock {
		logging.Debug("CloudWatch: Using MOCK (flag override)")
		return &MockCloudWatchClient{}, nil
	}

	// Check config file
	appCfg, err := config.LoadConfig()
	if err != nil {
		logging.Warn(fmt.Sprintf("Could not load config: %v, defaulting to mock", err))
		return &MockCloudWatchClient{}, nil
	}

	if appCfg.Mode == "mock" {
		logging.Debug("CloudWatch: Using MOCK (from config)")
		return &MockCloudWatchClient{}, nil
	}

	// Attempt to create real AWS client
	logging.Debug("CloudWatch: Attempting to create real AWS client")
	client, err := NewRealCloudWatchClient(ctx, cfg.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create real CloudWatch client: %w", err)
	}

	logging.Debug("CloudWatch: Using REAL AWS")
	return client, nil
}
