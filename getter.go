package signalfx

import "sync/atomic"

// Getter is an interface that is used by DataPoint. Get must return any kind of
// int, float, string, nil or pointer to those types. Any other type is invalid.
type Getter interface {
	Get() (interface{}, error)
}

type Subtracter interface {
	Getter
	Subtract(int64)
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
// types. If val is a pointer type, its value should not be changed, when in a
// Reporter, except within a PreReportCallback, for goroutine safety.
func Value(val interface{}) Getter {
	return vg{val}
}

type vg struct {
	v interface{}
}

func (v vg) Get() (interface{}, error) {
	return v.v, nil
}

/************************** Int32 **************************/

// Int32 satisfies the Getter interface using an atomic operation. Therefore it
// is safe to modify, when in a Reporter, outside the PreReportCallback as long
// as it is done so atomically.
type Int32 int32

// NewInt32 returns a new Int32 set to val
func NewInt32(val int32) *Int32 {
	ret := Int32(val)
	return &ret
}

// Get satisfies the Getter interface
func (v *Int32) Get() (interface{}, error) {
	return atomic.LoadInt32((*int32)(v)), nil
}

// Set the value using an atomic operation
func (v *Int32) Set(val int32) {
	atomic.StoreInt32((*int32)(v), val)
}

// Inc atomically adds delta to an Int32
func (v *Int32) Inc(delta int32) int32 {
	return atomic.AddInt32((*int32)(v), delta)
}

// Value atomically returns the value of an Int32
func (v *Int32) Value() int32 {
	return atomic.LoadInt32((*int32)(v))
}

func (v *Int32) Subtract(delta int64) {
	atomic.AddInt32((*int32)(v), int32(-delta))
}

/************************** Int64 **************************/

// Int64 satisfies the Getter interface using an atomic operation. Therefore it
// is safe to modify, when in a Reporter, outside the PreReportCallback as long
// as it is done so atomically.
type Int64 int64

// NewInt64 returns a new Int64 set to val
func NewInt64(val int64) *Int64 {
	ret := Int64(val)
	return &ret
}

// Get satisfies the Getter interface
func (v *Int64) Get() (interface{}, error) {
	return atomic.LoadInt64((*int64)(v)), nil
}

// Set the value using an atomic operation
func (v *Int64) Set(val int64) {
	atomic.StoreInt64((*int64)(v), val)
}

// Inc atomically adds delta to an Int64
func (v *Int64) Inc(delta int64) int64 {
	return atomic.AddInt64((*int64)(v), delta)
}

// Value atomically returns the value of an Int64
func (v *Int64) Value() int64 {
	return atomic.LoadInt64((*int64)(v))
}

func (v *Int64) Subtract(delta int64) {
	atomic.AddInt64((*int64)(v), -delta)
}

/************************* Uint32 **************************/

// Uint32 satisfies the Getter interface using an atomic operation. Therefore it
// is safe to modify, when in a Reporter, outside the PreReportCallback as long
// as it is done so atomically.
type Uint32 uint32

// NewUint32 returns a new Uint32 set to val
func NewUint32(val int32) *Uint32 {
	ret := Uint32(val)
	return &ret
}

// Get satisfies the Getter interface
func (v *Uint32) Get() (interface{}, error) {
	return atomic.LoadUint32((*uint32)(v)), nil
}

// Set the value using an atomic operation
func (v *Uint32) Set(val uint32) {
	atomic.StoreUint32((*uint32)(v), val)
}

// Inc atomically adds delta to a Uint32
func (v *Uint32) Inc(delta uint32) uint32 {
	return atomic.AddUint32((*uint32)(v), delta)
}

// Value atomically returns the value of a Uint32
func (v *Uint32) Value() uint32 {
	return atomic.LoadUint32((*uint32)(v))
}

// Subtract atomically subtracts delta from a Uint32
func (v *Uint32) Subtract(delta int64) {
	atomic.AddUint32((*uint32)(v), uint32(-delta))
}

/************************* Uint64 **************************/

// Uint64 satisfies the Getter interface using an atomic operation. Therefore it
// is safe to modify, when in a Reporter, outside the PreReportCallback as long
// as it is done so atomically.
type Uint64 uint64

// NewUint64 returns a new Uint64 set to val
func NewUint64(val int64) *Uint64 {
	ret := Uint64(val)
	return &ret
}

// Get satisfies the Getter interface
func (v *Uint64) Get() (interface{}, error) {
	return atomic.LoadUint64((*uint64)(v)), nil
}

// Set the value using an atomic operation
func (v *Uint64) Set(val uint64) {
	atomic.StoreUint64((*uint64)(v), val)
}

// Inc atomically adds delta to a Uint64
func (v *Uint64) Inc(delta uint64) uint64 {
	return atomic.AddUint64((*uint64)(v), delta)
}

// Value atomically returns the value of a Uint64
func (v *Uint64) Value() uint64 {
	return atomic.LoadUint64((*uint64)(v))
}

// Subtract atomically subtracts delta from a Uint32
func (v *Uint64) Subtract(delta int64) {
	atomic.AddUint64((*uint64)(v), uint64(-delta))
}
