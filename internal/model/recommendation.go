package model

// Recommendation describes optimisation advice for a single EC2 instance.
type Recommendation struct {
	InstanceID      string  `json:"instance_id"`
	InstanceType    string  `json:"instance_type"`
	State           string  `json:"state"`
	AvgCPU          float64 `json:"avg_cpu"`
	PeakCPU         float64 `json:"peak_cpu"`
	MonthlyCost     float64 `json:"monthly_cost"`
	HourlyCost      float64 `json:"hourly_cost"`
	Action          string  `json:"action"`           // e.g. "Downsize", "Upsize", "Keep as-is"
	SuggestedType   string  `json:"suggested_type"`   // e.g. "t3.nano"
	EstimatedSaving float64 `json:"estimated_saving"` // per month, rough estimate
	Reason          string  `json:"reason"`
}
