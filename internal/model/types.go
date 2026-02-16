package model

type EC2Instance struct {
	ID string `json:"id"`
	InstanceType string `json:"instance_type"`
	State string `json:"state"`
	Tags map[string]string `json:"tags"`
}

