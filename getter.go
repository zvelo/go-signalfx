package signalfx

import "sync/atomic"

// Getter is an interface that is used by DataPoint. It may return any kind of
// integer, float, string or nil, including pointers. Any other type is invalid.
type Getter interface {
	Get() (interface{}, error)
}

// The GetterFunc type is an adapter to allow the use of ordinary functions as
// DataPoint Getters. If f is a function with the appropriate signature,
// GetterFunc(f) is a Getter object that calls f. f() must return a value that
// is nil, any int type, any float type, a string, a pointer to any of those
// types or a Getter that returns any of those types.
type GetterFunc func() (interface{}, error)

// Get calls f()
func (f GetterFunc) Get() (interface{}, error) {
	return f()
}

// ValueGetter is a convenience function for making a value satisfy the Getter
// interface. It is especially useful with pointers. value must be nil, any int
// type, any float type, a string, a pointer to any of those types or a Getter
// that returns any of those types.
func ValueGetter(value interface{}) Getter {
	return valueGetter{value}
}

type valueGetter struct {
	value interface{}
}

func (v valueGetter) Get() (interface{}, error) {
	return v.value, nil
}

type Incrementer struct {
	value int64
}

func NewIncrementer(value int64) *Incrementer {
	return &Incrementer{value}
}

func (i *Incrementer) Set(value int64) {
	atomic.StoreInt64(&i.value, value)
}

func (i *Incrementer) Inc(delta int64) int64 {
	return atomic.AddInt64(&i.value, delta)
}

func (i *Incrementer) Get() (interface{}, error) {
	return i.Value(), nil
}

func (i *Incrementer) Value() int64 {
	return atomic.LoadInt64(&i.value)
}
