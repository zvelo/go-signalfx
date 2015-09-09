package signalfx

import (
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
