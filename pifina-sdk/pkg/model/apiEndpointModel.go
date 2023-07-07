package model

type ApiEndpointModel struct {
	Endpoint string `json:"endpoint"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
}
