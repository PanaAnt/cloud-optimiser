package model

// CPUSampleSeries represents a slice of CPU utilisation samples.
type CPUSampleSeries struct {
	InstanceID string
	Samples    []float64
}