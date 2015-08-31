package signalfx

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

type Counter struct {
	metric     string
	dimensions map[string]string
	value      uint64
}

var counterType = sfxproto.MetricType_COUNTER

func NewCounter(metric string, dimensions map[string]string, value uint64) *Counter {
	return &Counter{metric: metric, dimensions: dimensions, value: value}
}

func (c *Counter) Inc(delta uint64) uint64 {
	return atomic.AddUint64(&c.value, delta)
}

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

func (c *Counter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative counter should be impossible")
	}
	vv := uint64(v)
	// this is the canonical way to subtract a value atomically;
	// c.f. godoc atomic
	atomic.AddUint64(&c.value, ^(vv - 1))
}

type WrappedCounter struct {
	metric     string
	dimensions map[string]string
	value      Subtractor
}

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

func (c *WrappedCounter) DataPoint() *DataPoint {
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
		Type:       CounterType,
		Dimensions: c.dimensions,
		Value:      value,
	}
}

func (c *WrappedCounter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative counter should be impossible")
	}
	c.value.Subtract(v)
}
