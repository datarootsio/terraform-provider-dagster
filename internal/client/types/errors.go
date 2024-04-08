package types

import "fmt"

type ErrTeamNotFound struct {
	Name string
	Id   string
}

func (e *ErrTeamNotFound) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("team with name(%s) not found", e.Name)
	}

	if e.Id != "" {
		return fmt.Sprintf("team with id(%s) not found", e.Id)
	}

	return "team not found"
}

type ErrTeamAlreadyExists struct {
	Name string
}

func (e *ErrTeamAlreadyExists) Error() string {
	return fmt.Sprintf("team with name(%s) already exists", e.Name)

}

type ErrApi struct {
	Typename string
	Message  string
}

func (e *ErrApi) Error() string {
	return fmt.Sprintf("typename(%s): %s", e.Typename, e.Message)
}
