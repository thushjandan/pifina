package model

type DevPort struct {
	Name   string `json:"name"`
	PortId uint32 `json:"portId,omitempty"`
}
