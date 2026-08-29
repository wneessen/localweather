package apperror

import "fmt"

type CoordinateParsingError struct {
	Val string
	Err error
}

func (e *CoordinateParsingError) Error() string {
	return fmt.Sprintf("failed to parse coordinate value (%s): %s", e.Val, e.Err.Error())
}
