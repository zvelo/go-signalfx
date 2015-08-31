package signalfx

import (
	"math"
	"sync/atomic"
	"time"
)

// A Counter represents a Counter metric (i.e., its internal value is
// reset to zero by its PostReportHook, although this should not be
// meaningful to client code).
type Counter struct {
	metric     string
	dimensions map[string]string
	value      uint64
}

// NewCounter returns a new Counter with the specified parameters.  It
// does not copy the dimensions; client code should take care not to
// modify them in a goroutine-unsafe manner.
func NewCounter(
	metric string,
	dimensions map[string]string,
	value uint64,
) *Counter {
	return &Counter{metric: metric, dimensions: dimensions, value: value}
}

// Inc increments a Counter's internal state.  It is goroutine-safe.
func (c *Counter) Inc(delta uint64) uint64 {
	return atomic.AddUint64(&c.value, delta)
}

// DataPoint returns a DataPoint reflecting the Counter's internal
// state.  It does not reset that state, leaving that to
// PostReportHook.
func (c *Counter) DataPoint() *DataPoint {
	value := atomic.LoadUint64(&c.value)
	if value == 0 || value > math.MaxInt64 {
		return nil
	}
	return &DataPoint{
		Metric:     c.metric,
		Timestamp:  time.Now(),
		Type:       CounterType,
		Dimensions: c.dimensions,
		Value:      int64(value),
	}
}

// PostReportHook resets a Counter's internal state.  It does this
// intelligently, by subtracting the successfully-reported value from
// the state (this way, the Counter may have been incremented in
// between the create of the reported DataPoint and the reset, and the
// state will be consistent).  Calling PostReportHook with a negative
// value will result in a panic: counters may never take negative
// values.
func (c *Counter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative counter should be impossible")
	}
	vv := uint64(v)
	// this is the canonical way to subtract a value atomically;
	// c.f. godoc atomic
	atomic.AddUint64(&c.value, ^(vv - 1))
}

// A WrappedCounter wraps Subtractor—a type of Value which may be
// gotten from or subtracted from—in order to provide an easy-to-use
// way to monitor state elsewhere within an application.  Of note to
// client code, this means that the wrapped counter's value will be
// reset after each successful report.  Clients wishing to have a
// monotonically-increasing counter should instead use a
// WrappedCumulativeCounter.
type WrappedCounter struct {
	metric     string
	dimensions map[string]string
	value      Subtractor
}

// WrapCounter returns a WrappedCounter wrapping value.
func WrapCounter(
	metric string,
	dimensions map[string]string,
	value Subtractor,
) *WrappedCounter {
	return &WrappedCounter{
		metric:     metric,
		dimensions: dimensions,
		value:      value,
	}
}

// DataPoint returns a DataPoint representing the current state of the
// wrapped counter value.  If that current state is negative, then no
// DataPoint will be returned.
func (c *WrappedCounter) DataPoint() *DataPoint {
	gottenValue, err := c.value.Get()
	if err != nil {
		return nil
	}
	value, err := toInt64(gottenValue)
	if err != nil {
		return nil
	}
	if value < 0 {
		return nil
	}
	return &DataPoint{
		Metric:     c.metric,
		Timestamp:  time.Now(),
		Type:       CounterType,
		Dimensions: c.dimensions,
		Value:      value,
	}
}

// PostReportHook resets a counter by subtracting the
// successfully-reported value from its internal state.  Passing in a
// negative value will result in a panic, as a WrappedCounter's
// internal state may not be negative.
func (c *WrappedCounter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative counter should be impossible")
	}
	c.value.Subtract(v)
}
