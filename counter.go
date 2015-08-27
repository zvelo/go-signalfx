package signalfx

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

type Counter struct {
	Metric     string
	Dimensions map[string]string
	value      uint64
}

var counterType = sfxproto.MetricType_COUNTER

func (c *Counter) Inc(delta uint64) uint64 {
	return atomic.AddUint64(&c.value, delta)
}

func (c *Counter) dataPoint() *dataPoint {
	value := atomic.LoadUint64(&c.value)
	if value > math.MaxInt64 {
		return nil
	}
	return &dataPoint{
		Metric:     c.Metric,
		Timestamp:  time.Now(),
		Type:       CounterType,
		Dimensions: c.Dimensions,
		Value:      int64(value),
	}
}

func (c *Counter) reset(v int64) {
	if v < 0 {
		panic("negative counter should be impossible")
	}
	vv := uint64(v)
	// this is the canonical way to subtract a value atomically;
	// c.f. godoc atomic
	atomic.AddUint64(&c.value, ^(vv - 1))
}

type SetterCounter struct {
	Metric     string
	Dimensions map[string]string
	Value      Subtracter
}

func (c *SetterCounter) dataPoint() *dataPoint {
	gottenValue, err := c.Value.Get()
	if err != nil {
		return nil
	}
	value, err := toInt64(gottenValue)
	if err != nil {
		return nil
	}
	return &dataPoint{
		Metric:     c.Metric,
		Timestamp:  time.Now(),
		Type:       CounterType,
		Dimensions: c.Dimensions,
		Value:      value,
	}
}

func (c *SetterCounter) reset(v int64) {
	if v < 0 {
		panic("negative counter should be impossible")
	}
	c.Value.Subtract(v)
}
