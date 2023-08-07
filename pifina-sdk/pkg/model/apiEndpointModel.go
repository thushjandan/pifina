// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package model

type ApiEndpointModel struct {
	Endpoint string `json:"name"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
}
