package signalfx

import "fmt"

// ErrMarshal is returned when there is an error marshaling the DataPoints
type ErrMarshal error

// ErrContext is returned when the context is canceled or times out
type ErrContext error

// ErrPost is returned when there is an error posting data to signalfx
type ErrPost error

// ErrResponse is returned when there is an error reading the signalfx
// response
type ErrResponse error

// ErrStatus is returned when the signalfx response code isn't 200
type ErrStatus struct {
	Body       []byte
	StatusCode int
}

func (e *ErrStatus) Error() string {
	return fmt.Sprintf("%s: invalid status code: %d", e.Body, e.StatusCode)
}

// ErrJSON is returned when there is an error parsing the signalfx JSON
// response
type ErrJSON struct {
	Body []byte
}

func (e *ErrJSON) Error() string {
	return string(e.Body)
}

// ErrInvalidBody is returned when the signalfx response is anything other than
// "OK"
type ErrInvalidBody struct {
	Body string
}

func (e *ErrInvalidBody) Error() string {
	return e.Body
}
