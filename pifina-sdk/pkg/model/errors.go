package model

import "fmt"

type ErrNameNotFound struct {
	Entity string
	Msg    string
}

func (e *ErrNameNotFound) Error() string {
	return fmt.Sprintf("%s - Entity: %s", e.Msg, e.Entity)
}
