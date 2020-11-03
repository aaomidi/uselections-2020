package redis

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
	return fmt.Sprintf("redis error: %s. %v", e.cause, e.base)
}
