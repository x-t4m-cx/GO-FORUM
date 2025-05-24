package models

import "fmt"

type ErrNotFound struct {
	Entity string
	Id     int
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s with ID %d not found", e.Entity, e.Id)
}
