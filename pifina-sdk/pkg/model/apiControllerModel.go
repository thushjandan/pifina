// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package model

type ApiErrorMessage struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
