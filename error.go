package signalfx

import "fmt"

// Error is the only error type returned from Submit
type Error struct {
	OriginalError error
	Message       string
}

func (f *Error) Error() string {
	return fmt.Sprintf("%s: %s", f.Message, f.OriginalError.Error())
}

// newError provides a convenient way to create an Error
func newError(msg string, err error) *Error {
	return &Error{
		OriginalError: err,
		Message:       msg,
	}
}
