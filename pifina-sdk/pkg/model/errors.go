// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package model

import "fmt"

type ErrNameNotFound struct {
	Entity string
	Msg    string
}

type ErrNotReady struct {
	Msg string
}

func (e *ErrNameNotFound) Error() string {
	return fmt.Sprintf("%s - Entity: %s", e.Msg, e.Entity)
}

func (e *ErrNotReady) Error() string {
	return e.Msg
}
