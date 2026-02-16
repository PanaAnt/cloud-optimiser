package awsclient

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
)

func CanUseRealAWS(ctx context.Context, profile string) error {
	opts := []func(*config.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	_, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("AWS config unavailable: %w", err)
	}

	return nil
}
