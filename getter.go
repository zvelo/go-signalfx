package signalfx

import "sync/atomic"

// Getter is an interface that is used by DataPoint. Get must return any kind of
// int, float, string, nil or pointer to those types. Any other type is invalid.
type Getter interface {
	Get() (interface{}, error)
}

// The GetterFunc type is an adapter to allow the use of ordinary functions as
// DataPoint Getters. If f is a function with the appropriate signature,
// GetterFunc(f) is a Getter object that calls f. f() must return a value that
// is any type of int, float, string, nil or pointer to those types.
type GetterFunc func() (interface{}, error)

// Get calls f()
func (f GetterFunc) Get() (interface{}, error) {
	return f()
}

// Value is a convenience function for making a value satisfy the Getter
// interface. val can be any type of int, float, string, nil or pointer to those
// types. If val is a pointer type, its value should not be changed, unless
// atomically, when in a Reporter, except within a PreReportCallback, for
// goroutine safety.
func Value(val interface{}) Getter {
	return vg{val}
}

type vg struct {
	v interface{}
}

func (v vg) Get() (interface{}, error) {
	return v.v, nil
}

// Inc is an incrementer object that satisfies the Getter interface. All
// operations on it are goroutine safe.
type Inc struct {
	// use a struct instead of typing on int64 to ensure goroutine safety
	v int64
}

// NewInc returns a new Inc initialized to val
func NewInc(val int64) *Inc {
	return &Inc{val}
}

// Set the value of the Inc
func (i *Inc) Set(val int64) {
	atomic.StoreInt64(&i.v, val)
}

// Inc adds delta to the existing value of the Inc
func (i *Inc) Inc(delta int64) int64 {
	return atomic.AddInt64(&i.v, delta)
}

// Get returns the value of the Inc, satisfies the Getter interface
func (i *Inc) Get() (interface{}, error) {
	return i.Value(), nil
}

// Value returns the value of the Inc
func (i *Inc) Value() int64 {
	return atomic.LoadInt64(&i.v)
}
