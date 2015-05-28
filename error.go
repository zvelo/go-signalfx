package signalfx

import "fmt"

// sfxError is the only error type returned from Submit
type sfxError struct {
	OriginalError error
	Message       string
}

func (f *sfxError) Error() string {
	return fmt.Sprintf("%s: %s", f.Message, f.OriginalError.Error())
}

// newError provides a convenient way to create an sfxError
func newError(msg string, err error) *sfxError {
	return &sfxError{
		OriginalError: err,
		Message:       msg,
	}
}
