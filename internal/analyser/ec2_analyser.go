package analyser

import (
	"context"
	"fmt"

	"github.com/PanaAnt/cloud-optimiser/internal/awsclient"
	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

// Thresholds for optimization rules
const (
	lowAvgCPUThreshold     = 20.0 // below this = underutilized
	lowPeakCPUThreshold    = 40.0 // if peak also low, strong downsize candidate
	highAvgCPUThreshold    = 75.0 // above this = heavily utilized
	idleCPUThreshold       = 5.0  // below this considered idle sample
	downsizeSavingEstimate = 0.3  // 30% estimated savings for downsize
)

// AnalyseInstances fetches metrics + costs and generates recommendations.
func AnalyseInstances(
	ctx context.Context,
	instances []model.EC2Instance,
	cw awsclient.CloudWatchClient,
	ce awsclient.CostExplorerClient,
	metricHours int,
	costDays int,
) ([]model.Recommendation, error) {
	var recs []model.Recommendation

	for _, inst := range instances {
		if inst.State == "terminated" {
			continue
		}

		// 1) Fetch CPU metrics
		cpuSeries, err := cw.GetCpuUtilisation(ctx, inst.ID, metricHours)
		if err != nil {
			
			recs = append(recs, model.Recommendation{
				InstanceID:   inst.ID,
				InstanceType: inst.InstanceType,
				State:        inst.State,
				Action:       "Unknown",
				Reason:       fmt.Sprintf("Failed to load CPU metrics: %v", err),
			})
			continue
		}

		avgCPU := average(cpuSeries.Samples)
		peakCPU := max(cpuSeries.Samples)
		idleRatio := fractionBelow(cpuSeries.Samples, idleCPUThreshold)

		// Fetch cost data
		cost, err := ce.GetInstanceCost(ctx, inst.ID, costDays)
		if err != nil {
			// Cost failure: still produce recommendation without savings
			cost = model.CostData{
				InstanceID:  inst.ID,
				MonthlyCost: 0,
				HourlyCost:  0,
			}
		}

		// Apply optimisation rules
		action := "Keep as-is"
		suggestedType := ""
		estimatedSaving := 0.0
		reason := ""

		if len(cpuSeries.Samples) == 0 {
			action = "Review / Potentially Stop"
			reason = "No CPU data available; instance may be idle or not sending metrics."
		} else if avgCPU < lowAvgCPUThreshold && peakCPU < lowPeakCPUThreshold {
			action = "Downsize"
			suggestedType = downsizeInstanceType(inst.InstanceType)
			if cost.MonthlyCost > 0 && suggestedType != inst.InstanceType {
				estimatedSaving = cost.MonthlyCost * downsizeSavingEstimate
			}
			reason = fmt.Sprintf("Average CPU %.1f%%, peak %.1f%%, idle %.0f%% of samples; strong downsize candidate.",
				avgCPU, peakCPU, idleRatio*100)
		} else if avgCPU > highAvgCPUThreshold {
			action = "Upsize / Scale out"
			suggestedType = upsizeInstanceType(inst.InstanceType)
			reason = fmt.Sprintf("Average CPU %.1f%% over %d hours; instance appears heavily utilized.",
				avgCPU, metricHours)
		} else {
			action = "Keep as-is"
			reason = fmt.Sprintf("Average CPU %.1f%%, peak %.1f%%; utilization appears reasonable.",
				avgCPU, peakCPU)
		}

		recs = append(recs, model.Recommendation{
			InstanceID:      inst.ID,
			InstanceType:    inst.InstanceType,
			State:           inst.State,
			AvgCPU:          avgCPU,
			PeakCPU:         peakCPU,
			MonthlyCost:     cost.MonthlyCost,
			HourlyCost:      cost.HourlyCost,
			Action:          action,
			SuggestedType:   suggestedType,
			EstimatedSaving: estimatedSaving,
			Reason:          reason,
		})
	}

	return recs, nil
}

// --- Helper functions ---

// average calculates the mean of a slice of floats
func average(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range xs {
		sum += v
	}
	return sum / float64(len(xs))
}

// max returns the maximum value in a slice of floats
func max(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	m := xs[0]
	for _, v := range xs {
		if v > m {
			m = v
		}
	}
	return m
}

// fractionBelow returns the fraction of samples strictly below threshold
func fractionBelow(xs []float64, threshold float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	count := 0
	for _, v := range xs {
		if v < threshold {
			count++
		}
	}
	return float64(count) / float64(len(xs))
}

// downsizeInstanceType suggests a smaller instance type
// TODO: Replace with JSON catalog for comprehensive instance type support
func downsizeInstanceType(current string) string {
	switch current {
	case "t3.micro":
		return "t3.nano"
	case "t3.small":
		return "t3.micro"
	case "t3.medium":
		return "t3.small"
	case "m5.large":
		return "m5.medium"
	case "m5.xlarge":
		return "m5.large"
	case "m5.2xlarge":
		return "m5.xlarge"
	default:
		return current // unknown: keep same type
	}
}

// upsizeInstanceType suggests a larger instance type
// TODO: Replace with JSON catalog for comprehensive instance type support
func upsizeInstanceType(current string) string {
	switch current {
	case "t3.nano":
		return "t3.micro"
	case "t3.micro":
		return "t3.small"
	case "t3.small":
		return "t3.medium"
	case "m5.medium":
		return "m5.large"
	case "m5.large":
		return "m5.xlarge"
	case "m5.xlarge":
		return "m5.2xlarge"
	default:
		return current // unknown: keep same type
	}
}

