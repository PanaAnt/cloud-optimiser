package model

type CostData struct {
	InstanceID string `json:"instance_id"`
	MonthlyCost float64 `json:"monthly_cost"`
	HourlyCost float64 `json:"hourly_cost"`

}