package awsclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	costexplorer "github.com/aws/aws-sdk-go-v2/service/costexplorer"
	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"

	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type RealCostExplorer struct {
	ce *costexplorer.Client
}

// NewRealCostExplorer creates a real AWS Cost Explorer client with optional profile
func NewRealCostExplorer(ctx context.Context, profile string) (*RealCostExplorer, error) {
	opts := []func(*config.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &RealCostExplorer{
		ce: costexplorer.NewFromConfig(cfg),
	}, nil
}

// IsMock returns false indicating real AWS client
func (r *RealCostExplorer) IsMock() bool {
	return false
}

// GetInstanceCost retrieves cost data from AWS Cost Explorer for a given instance
func (r *RealCostExplorer) GetInstanceCost(ctx context.Context, instanceID string, days int) (model.CostData, error) {
	end := time.Now().UTC()
	start := end.Add(-time.Duration(days) * 24 * time.Hour)

	input := &costexplorer.GetCostAndUsageInput{
		Metrics:     []string{"UnblendedCost"},
		Granularity: ceTypes.GranularityDaily,
		TimePeriod: &ceTypes.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		GroupBy: []ceTypes.GroupDefinition{
			{
				Type: ceTypes.GroupDefinitionTypeDimension,
				Key:  aws.String("RESOURCE_ID"),
			},
		},
	}

	resp, err := r.ce.GetCostAndUsage(ctx, input)
	if err != nil {
		return model.CostData{}, fmt.Errorf("GetCostAndUsage failed: %w", err)
	}

	var total float64 = 0

	// Walk through cost results to find matching instance
	for _, result := range resp.ResultsByTime {
		for _, group := range result.Groups {
			keys := group.Keys
			if len(keys) > 0 && keys[0] == instanceID {
				metric, ok := group.Metrics["UnblendedCost"]
				if ok && metric.Amount != nil {
					val, err := strconv.ParseFloat(aws.ToString(metric.Amount), 64)
					if err != nil {
						continue // Skip invalid amounts
					}
					total += val
				}
			}
		}
	}

	// Convert total cost (for the window) into hourly/monthly estimates
	hours := float64(days * 24)
	hourly := total / hours
	monthly := hourly * 24 * 30 // approximate 30-day month

	return model.CostData{
		InstanceID:  instanceID,
		MonthlyCost: monthly,
		HourlyCost:  hourly,
	}, nil
}

