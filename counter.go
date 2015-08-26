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

func (c *Counter) dataPoint(
	metricPrefix string,
	dimensions []*sfxproto.Dimension,
) *sfxproto.DataPoint {
	timestamp := time.Now().UnixNano() / 1000000
	protoDims := make(
		[]*sfxproto.Dimension,
		len(dimensions),
		len(dimensions)+len(c.Dimensions))
	for i, v := range dimensions {
		protoDims[i] = v
	}
	for k, v := range c.Dimensions {
		// have to copy the values, since these are stored as
		// pointers…
		var dk, dv string
		dk = k
		dv = v
		protoDims = append(
			protoDims,
			&sfxproto.Dimension{Key: &dk, Value: &dv},
		)
	}
	prefixedMetric := metricPrefix + c.Metric
	value := atomic.LoadUint64(&c.value)
	if value > math.MaxInt64 {
		return nil
	}
	iv := int64(value)
	return &sfxproto.DataPoint{
		Metric:     &prefixedMetric,
		Timestamp:  &timestamp,
		MetricType: &counterType,
		Dimensions: protoDims,
		Value:      &sfxproto.Datum{IntValue: &iv},
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
