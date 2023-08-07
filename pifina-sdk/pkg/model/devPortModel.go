// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package model

type DevPort struct {
	Name   string `json:"name"`
	PortId uint32 `json:"portId,omitempty"`
}
