package signalfx

import "sync/atomic"

// Getter is an interface that is used by Metric. It may return any kind of
// integer, float, string or nil, including pointers. Any other type is invalid.
type Getter interface {
	Get() (interface{}, error)
}

// The GetterFunc type is an adapter to allow the use of ordinary functions as
// Metric Getters. If f is a function with the appropriate signature,
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

type Incrementer int64

func (i *Incrementer) Set(value int64) {
	atomic.StoreInt64((*int64)(i), value)
}

func (i *Incrementer) Inc(delta int64) int64 {
	return atomic.AddInt64((*int64)(i), delta)
}

func (i *Incrementer) Get() (interface{}, error) {
	return atomic.LoadInt64((*int64)(i)), nil
}
