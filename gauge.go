package signalfx

import (
	"math"
	"sync/atomic"
	"time"
)

// A Gauge represents a metric which tracks a single value.  This
// value may be positive or negative, and may increase or decrease
// over time.  It neither copies nor modifies its dimension; client
// code should ensure that it does not modify them in a thread-unsafe
// manner.  Unlike counters and cumulative counters, gauges are always
// reported.
type Gauge struct {
	metric     string
	dimensions map[string]string
	value      int64
}

// NewGauge returns a new Gauge with the indicated initial state.
func NewGauge(metric string, dimensions map[string]string, value int64) *Gauge {
	return &Gauge{metric: metric, dimensions: dimensions, value: value}
}

// Record sets a gauge's internal state to the indicated value.
func (g *Gauge) Record(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// DataPoint returns a DataPoint reflecting the Gauge's internal state
// at the current point in time.
func (g *Gauge) DataPoint() *DataPoint {
	return &DataPoint{
		Metric:     g.metric,
		Timestamp:  time.Now(),
		Type:       GaugeType,
		Dimensions: g.dimensions,
		Value:      atomic.LoadInt64(&g.value),
	}
}

// A WrappedGauge wraps a Getter elsewhere in memory.
type WrappedGauge struct {
	metric     string
	dimensions map[string]string
	value      Getter
}

// WrapGauge wraps a value in memory.  It neither copies nor updates
// its dimensions; client code should take care not to modify them in
// a goroutine-unsafe manner.
func WrapGauge(
	metric string,
	dimensions map[string]string,
	value Getter,
) *WrappedGauge {
	return &WrappedGauge{
		metric:     metric,
		dimensions: dimensions,
		value:      value,
	}
}

// DataPoint returns a DataPoint reflecting the current value of the
// WrappedGauge.
func (c *WrappedGauge) DataPoint() *DataPoint {
	gottenValue, err := c.value.Get()
	if err != nil {
		return nil
	}
	value, err := toInt64(gottenValue)
	if err != nil {
		return nil
	}
	return &DataPoint{
		Metric:     c.metric,
		Timestamp:  time.Now(),
		Type:       GaugeType,
		Dimensions: c.dimensions,
		Value:      value,
	}
}

// A StableGauge is the same as a Guage, but will not report the same value
// multiple times sequentially.
type StableGauge struct {
	gauge     *Gauge
	prevValue int64
}

// NewStableGauge returns a new StableGauge with the indicated initial state.
func NewStableGauge(metric string, dimensions map[string]string, value int64) *StableGauge {
	return &StableGauge{gauge: NewGauge(metric, dimensions, value), prevValue: math.MaxInt64}
}

// Record sets a gauge's internal state to the indicated value.
func (g *StableGauge) Record(value int64) {
	g.gauge.Record(value)
}

// DataPoint returns a DataPoint reflecting the StableGauge's internal state
// at the current point in time, if it differs since the last call.
func (g *StableGauge) DataPoint() *DataPoint {
	dp := g.gauge.DataPoint()
	if dp.Value == atomic.LoadInt64(&g.prevValue) {
		// Value is the same, don't report it
		return nil
	}
	atomic.StoreInt64(&g.prevValue, dp.Value)
	return dp
}
