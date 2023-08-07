// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package model

type NeoHostDeviceList struct {
	Id      int                 `json:"id"`
	Error   NeoHostError        `json:"error"`
	Results []NeoHostDeviceItem `json:"result"`
}

type NeoHostError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

type NeoHostDeviceItem struct {
	Name  string                  `json:"name"`
	UID   string                  `json:"uid"`
	Ports []NeoHostDeviceItemPort `json:"ports"`
}

type NeoHostDeviceItemPort struct {
	UID               string                              `json:"uid"`
	IbDevice          string                              `json:"ibDevice"`
	Number            int                                 `json:"number"`
	PhysicalFunctions []NeoHostDeviceItemPhysicalFunction `json:"physicalFunctions"`
}

type NeoHostDeviceItemPhysicalFunction struct {
	NetworkInterfaces []string `json:"networkInterfaces"`
	UID               string   `json:"uid"`
}
