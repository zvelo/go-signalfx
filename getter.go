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

// Value is a convenience function for making a value satisfy the Getter
// interface. It is especially useful with pointers. value must be nil, any int
// type, any float type, a string, a pointer to any of those types or a Getter
// that returns any of those types.
func Value(val interface{}) Getter {
	return vg{val}
}

type vg struct {
	v interface{}
}

func (v vg) Get() (interface{}, error) {
	return v.v, nil
}

type Inc struct {
	// use a struct instead of typing on int64 to ensure goroutine safety
	v int64
}

func NewInc(val int64) *Inc {
	return &Inc{val}
}

func (i *Inc) Set(val int64) {
	atomic.StoreInt64(&i.v, val)
}

func (i *Inc) Inc(delta int64) int64 {
	return atomic.AddInt64(&i.v, delta)
}

func (i *Inc) Get() (interface{}, error) {
	return i.Value(), nil
}

func (i *Inc) Value() int64 {
	return atomic.LoadInt64(&i.v)
}
