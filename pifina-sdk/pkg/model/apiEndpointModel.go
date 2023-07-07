package model

type ApiEndpointModel struct {
	Endpoint string `json:"name"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
}
