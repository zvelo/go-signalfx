package signalfx

import "fmt"

// sfError is the only error type returned from Submit
type sfError struct {
	OriginalError error
	Message       string
}

func (f *sfError) Error() string {
	return fmt.Sprintf("%s: %s", f.Message, f.OriginalError.Error())
}

// newError provides a convenient way to create an sfError
func newError(msg string, err error) *sfError {
	return &sfError{
		OriginalError: err,
		Message:       msg,
	}
}
