package types

import "fmt"

type ErrNotFound struct {
	What  string
	Key   string
	Value string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s %s=%s not found", e.What, e.Key, e.Value)
}

type ErrAlreadyExists struct {
	What  string
	Key   string
	Value string
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s %s=%s already exists", e.What, e.Key, e.Value)
}

type ErrApi struct {
	Typename string
	Message  string
}

func (e *ErrApi) Error() string {
	return fmt.Sprintf("typename(%s): %s", e.Typename, e.Message)
}
