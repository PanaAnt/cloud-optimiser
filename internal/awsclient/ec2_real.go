package awsclient

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

type RealClient struct {
	ec2Client *ec2.Client
}

// NewRealClient creates a real AWS EC2 client with optional profile
func NewRealClient(ctx context.Context, profile string) (*RealClient, error) {
	opts := []func(*config.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &RealClient{
		ec2Client: ec2.NewFromConfig(cfg),
	}, nil
}

// IsMock returns false indicating real AWS client
func (r *RealClient) IsMock() bool {
	return false
}

// ListInstances retrieves all EC2 instances from AWS
func (r *RealClient) ListInstances(ctx context.Context) ([]model.EC2Instance, error) {
	input := &ec2.DescribeInstancesInput{}

	result, err := r.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed: %w", err)
	}

	var instances []model.EC2Instance
	for _, res := range result.Reservations {
		for _, inst := range res.Instances {
			tags := map[string]string{}
			for _, tag := range inst.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}

			instance := model.EC2Instance{
				ID:           aws.ToString(inst.InstanceId),
				InstanceType: string(inst.InstanceType),
				State:        string(inst.State.Name),
				Tags:         tags,
			}

			instances = append(instances, instance)
		}
	}
	return instances, nil
}
