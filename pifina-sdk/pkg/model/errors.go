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
