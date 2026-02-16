package awsclient

import (
	"context"
	"fmt"

	"github.com/PanaAnt/cloud-optimiser/internal/config"
	"github.com/PanaAnt/cloud-optimiser/internal/logging"
	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type EC2Client interface {
	ListInstances(ctx context.Context) ([]model.EC2Instance, error)
	IsMock() bool
}

type Config struct {
	UseMock bool
	Profile string
}

// New creates an EC2 client based on the provided configuration.
// Returns an error if real AWS client creation fails and mock mode is not enabled.
func New(ctx context.Context, cfg Config) (EC2Client, error) {
	// Forced mock via flag
	if cfg.UseMock {
		logging.Debug("EC2: Using MOCK (flag override)")
		return &MockClient{}, nil
	}

	// Check config file
	appCfg, err := config.LoadConfig()
	if err != nil {
		logging.Warn(fmt.Sprintf("Could not load config: %v, defaulting to mock", err))
		return &MockClient{}, nil
	}

	if appCfg.Mode == "mock" {
		logging.Debug("EC2: Using MOCK (from config)")
		return &MockClient{}, nil
	}

	// Attempt to create real AWS client
	logging.Debug("EC2: Attempting to create real AWS client")
	real, err := NewRealClient(ctx, cfg.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create real EC2 client: %w", err)
	}

	logging.Debug("EC2: Using REAL AWS")
	return real, nil
}
