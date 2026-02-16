package awsclient

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type RealCloudWatchClient struct {
	cw *cloudwatch.Client
}

// NewRealCloudWatchClient creates a real AWS CloudWatch client with optional profile
func NewRealCloudWatchClient(ctx context.Context, profile string) (*RealCloudWatchClient, error) {
	opts := []func(*config.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &RealCloudWatchClient{
		cw: cloudwatch.NewFromConfig(cfg),
	}, nil
}

// IsMock returns false indicating real AWS client
func (r *RealCloudWatchClient) IsMock() bool {
	return false
}

// GetCpuUtilisation retrieves CPU utilisation metrics from CloudWatch for a given instance
func (r *RealCloudWatchClient) GetCpuUtilisation(
	ctx context.Context,
	instanceID string,
	hours int,
) (model.CPUSampleSeries, error) {
	end := time.Now().UTC()
	start := end.Add(-time.Duration(hours) * time.Hour)
	metricID := "cpuMetric"

	input := &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(start),
		EndTime:   aws.Time(end),
		MetricDataQueries: []cloudwatchtypes.MetricDataQuery{
			{
				Id: aws.String(metricID),
				MetricStat: &cloudwatchtypes.MetricStat{
					Metric: &cloudwatchtypes.Metric{
						Namespace:  aws.String("AWS/EC2"),
						MetricName: aws.String("CPUUtilization"),
						Dimensions: []cloudwatchtypes.Dimension{
							{
								Name:  aws.String("InstanceId"),
								Value: aws.String(instanceID),
							},
						},
					},
					Period: aws.Int32(300), // 5 minute intervals
					Stat:   aws.String("Average"),
				},
			},
		},
	}

	result, err := r.cw.GetMetricData(ctx, input)
	if err != nil {
		return model.CPUSampleSeries{}, fmt.Errorf("GetMetricData failed: %w", err)
	}

	var samples []float64
	for _, res := range result.MetricDataResults {
		if aws.ToString(res.Id) == metricID {
			samples = append(samples, res.Values...)
		}
	}

	return model.CPUSampleSeries{
		InstanceID: instanceID,
		Samples:    samples,
	}, nil
}



