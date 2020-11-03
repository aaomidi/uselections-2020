package telegram

import "fmt"

type Error struct {
	base  error
	cause string
}

func NewError(err error, cause string) Error {
	return Error{
		base:  err,
		cause: cause,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("telegram error: %s. %v", e.cause, e.base)
}
