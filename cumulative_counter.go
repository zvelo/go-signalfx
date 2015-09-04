package signalfx

import (
	"math"
	"sync/atomic"
	"time"
)

// A CumulativeCounter represents a cumulative counter, that is a
// counter whose internal state monotonically increases, and which
// never reports a value it has previously reported.  It may be useful
// in order to track the value of counters over which one has no
// control.
type CumulativeCounter struct {
	metric               string
	dimensions           map[string]string
	value, previousValue uint64
}

// NewCumulativeCounter returns a newly-created CumulativeCounter with
// the indicated initial state.  It neither copies nor modifies the
// dimensions; client code should not modify dimensions in a
// goroutine-unsafe manner.
func NewCumulativeCounter(metric string, dimensions map[string]string, value uint64) *CumulativeCounter {
	return &CumulativeCounter{metric: metric, dimensions: dimensions, value: value}
}

// Sample updates the CumulativeCounter's internal state.
func (cc *CumulativeCounter) Sample(delta uint64) {
	atomic.StoreUint64(&cc.value, delta)
}

// DataPoint returns a DataPoint reflecting the internal state of the
// CumulativeCounter.  If that state has previously been successfully
// reported (as recorded by PostReportHook), it will simply return
// nil.
func (cc *CumulativeCounter) DataPoint() *DataPoint {
	previous := atomic.LoadUint64(&cc.previousValue)
	value := atomic.LoadUint64(&cc.value)
	if value == previous {
		return nil
	}
	if value > math.MaxInt64 {
		return nil
	}
	return &DataPoint{
		Metric:     cc.metric,
		Timestamp:  time.Now(),
		Type:       CumulativeCounterType,
		Dimensions: cc.dimensions,
		Value:      int64(value),
	}
}

// PostReportHook records a reported value of a CumulativeCounter.  It
// will do the right thing if a yet-higher value has been reported in
// the interval since DataPoint was called.  Calling PostReportHook
// with a negative value will result in a panic: counters may never
// take negative values.
//
// In the normal case, PostReportHook should only be called by
// Reporter.Report.  Its argument must always be the value of a
// DataPoint previously returned by CumulativeCounter.DataPoint.
func (cc *CumulativeCounter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative cumulative counter should be impossible")
	}
	vv := uint64(v)
	prev := atomic.LoadUint64(&cc.previousValue)
	if vv <= prev {
		return
	}
	for !atomic.CompareAndSwapUint64(&cc.previousValue, prev, vv) {
		prev = atomic.LoadUint64(&cc.previousValue)
		if vv <= prev {
			return
		}
	}
}

// A WrappedCumulativeCounter wraps a value elsewhere in memory.  That
// value must monotonically increase.
type WrappedCumulativeCounter struct {
	metric               string
	dimensions           map[string]string
	wrappedValue         Getter
	value, previousValue uint64
}

// WrapCumulativeCounter wraps a cumulative counter elsewhere in
// memory, returning a newly-allocated WrappedCumulativeCounter.
// Dimensions are neither copied nor modified; client code should take
// care not to modify them in a goroutine-unsafe manner.
func WrapCumulativeCounter(
	metric string,
	dimensions map[string]string,
	value Getter,
) *WrappedCumulativeCounter {
	return &WrappedCumulativeCounter{
		metric:       metric,
		dimensions:   dimensions,
		wrappedValue: value,
	}
}

// DataPoint returns a DataPoint reflecting the internal state of the
// WrappedCumulativeCounter at a particular point in time.  It will
// return nil of the counter's value or higher has been previously
// reported.
func (cc *WrappedCumulativeCounter) DataPoint() *DataPoint {
	previous := atomic.LoadUint64(&cc.previousValue)
	gottenValue, err := cc.wrappedValue.Get()
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
	if uint64(value) == previous {
		return nil
	}
	return &DataPoint{
		Metric:     cc.metric,
		Timestamp:  time.Now(),
		Type:       CumulativeCounterType,
		Dimensions: cc.dimensions,
		Value:      int64(value),
	}
}

// PostReportHook records that a particular value has been
// successfully reported.  If that value is negative, it will panic,
// as cumulative counter values may never be negative.  It is
// goroutine-safe.
func (cc *WrappedCumulativeCounter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative cumulative counter should be impossible")
	}
	vv := uint64(v)
	prev := atomic.LoadUint64(&cc.previousValue)
	if vv <= prev {
		return
	}
	for !atomic.CompareAndSwapUint64(&cc.previousValue, prev, vv) {
		prev = atomic.LoadUint64(&cc.previousValue)
		if vv <= prev {
			return
		}
	}
}
