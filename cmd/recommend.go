package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/PanaAnt/cloud-optimiser/internal/analyser"
	"github.com/PanaAnt/cloud-optimiser/internal/awsclient"
	"github.com/PanaAnt/cloud-optimiser/internal/config"
	"github.com/PanaAnt/cloud-optimiser/internal/logging"
	"github.com/PanaAnt/cloud-optimiser/internal/model"
)

var (
	metricHours int
	costDays    int

	sortBy       string
	outputFormat string
	onlyDownsize bool
	onlyUpsize   bool
	stateFilter  string
	minCPU       float64
)

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Analyse EC2 instances and print optimization recommendations",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// 1. Load config & resolve initial mode
		cfg, err := config.LoadConfig()
		if err != nil {
			logging.Warn(fmt.Sprintf("Could not load config: %v, defaulting to MOCK mode", err))
			cfg = config.AppConfig{Mode: "mock"}
		}

		useMockMode := useMock || cfg.Mode == "mock"

		// 2. AWS readiness check (only if not already in mock mode)
		if !useMockMode {
			if err := awsclient.CanUseRealAWS(ctx, awsProfile); err != nil {
				fmt.Println("AWS unavailable – switching to MOCK mode.")
				logging.DebugErr("AWS readiness check failed", err)
				useMockMode = true
			}
		}

		// 3. Display mode
		fmt.Println("MODE:", map[bool]string{true: "Mock", false: "Real AWS"}[useMockMode])
		fmt.Println()

		// 4. Create AWS clients
		clientCfg := awsclient.Config{
			UseMock: useMockMode,
			Profile: awsProfile,
		}

		ec2Client, err := awsclient.New(ctx, clientCfg)
		if err != nil {
			fmt.Printf("Failed to create EC2 client: %v\n", err)
			logging.DebugErr("EC2 client creation failed", err)
			return
		}

		cwClient, err := awsclient.NewCloudWatch(ctx, clientCfg)
		if err != nil {
			fmt.Printf("Failed to create CloudWatch client: %v\n", err)
			logging.DebugErr("CloudWatch client creation failed", err)
			return
		}

		ceClient, err := awsclient.NewCostExplorer(ctx, clientCfg)
		if err != nil {
			fmt.Printf("Failed to create Cost Explorer client: %v\n", err)
			logging.DebugErr("Cost Explorer client creation failed", err)
			return
		}

		// 5. Fetch EC2 instances
		instances, err := ec2Client.ListInstances(ctx)
		if err != nil {
			fmt.Printf("Failed to list instances: %v\n", err)
			logging.DebugErr("EC2 ListInstances failed", err)
			return
		}

		if len(instances) == 0 {
			fmt.Println("No EC2 instances found.")
			return
		}

		// 6. Run analysis
		recs, err := analyser.AnalyseInstances(
			ctx,
			instances,
			cwClient,
			ceClient,
			metricHours,
			costDays,
		)
		if err != nil {
			fmt.Printf("Analysis failed: %v\n", err)
			logging.DebugErr("AnalyseInstances failed", err)
			return
		}

		// 7. Filter + sort
		recs = applyFilters(recs)
		sortRecommendations(recs)

		// 8. Output
		if outputFormat == "json" {
			outputJSON(recs)
		} else {
			outputTable(recs)
		}
		
		// 9. Indicate if using mock data
		if ec2Client.IsMock() {
			fmt.Println("\n[Note: Using mock data]")
		}
	},
}

func init() {
	rootCmd.AddCommand(recommendCmd)

	recommendCmd.Flags().IntVar(&metricHours, "metric-hours", 24, "Hours of CPU metrics to analyze")
	recommendCmd.Flags().IntVar(&costDays, "cost-days", 30, "Days of cost data to analyze")
	recommendCmd.Flags().StringVar(&sortBy, "sort", "none", "Sort by: cpu | cost | savings")
	recommendCmd.Flags().StringVar(&outputFormat, "output", "table", "Output format: table | json")
	recommendCmd.Flags().BoolVar(&onlyDownsize, "only-downsize", false, "Show only downsize recommendations")
	recommendCmd.Flags().BoolVar(&onlyUpsize, "only-upsize", false, "Show only upsize recommendations")
	recommendCmd.Flags().StringVar(&stateFilter, "state", "", "Filter by instance state (e.g., running, stopped)")
	recommendCmd.Flags().Float64Var(&minCPU, "min-cpu", 0, "Minimum average CPU threshold")
}

// applyFilters filters recommendations based on command line flags
func applyFilters(recs []model.Recommendation) []model.Recommendation {
	out := []model.Recommendation{}
	for _, r := range recs {
		if stateFilter != "" && r.State != stateFilter {
			continue
		}
		if onlyDownsize && r.Action != "Downsize" {
			continue
		}
		if onlyUpsize && r.Action != "Upsize / Scale out" {
			continue
		}
		if r.AvgCPU < minCPU {
			continue
		}
		out = append(out, r)
	}
	return out
}

// sortRecommendations sorts recommendations based on the sort flag
func sortRecommendations(recs []model.Recommendation) {
	switch sortBy {
	case "cpu":
		sort.Slice(recs, func(i, j int) bool { return recs[i].AvgCPU > recs[j].AvgCPU })
	case "cost":
		sort.Slice(recs, func(i, j int) bool { return recs[i].MonthlyCost > recs[j].MonthlyCost })
	case "savings":
		sort.Slice(recs, func(i, j int) bool { return recs[i].EstimatedSaving > recs[j].EstimatedSaving })
	}
}

// outputJSON outputs recommendations in JSON format
func outputJSON(recs []model.Recommendation) {
	b, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

// outputTable outputs recommendations in table format
func outputTable(recs []model.Recommendation) {
	w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tSTATE\tCPU(avg)\tCPU(peak)\tCOST/mo\tACTION\tNEW TYPE\tSAVING\tREASON")
	for _, r := range recs {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%.1f%%\t%.1f%%\t$%.2f\t%s\t%s\t$%.2f\t%s\n",
			r.InstanceID,
			r.InstanceType,
			r.State,
			r.AvgCPU,
			r.PeakCPU,
			r.MonthlyCost,
			r.Action,
			r.SuggestedType,
			r.EstimatedSaving,
			r.Reason,
		)
	}
	w.Flush()
}
