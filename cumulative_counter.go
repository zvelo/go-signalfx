package signalfx

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

type CumulativeCounter struct {
	Metric               string
	Dimensions           map[string]string
	value, previousValue uint64
}

var cumulativeCounterType = sfxproto.MetricType_CUMULATIVE_COUNTER

func (cc *CumulativeCounter) dataPoint(
	metricPrefix string,
	dimensions []*sfxproto.Dimension,
) *sfxproto.DataPoint {
	previous := atomic.LoadUint64(&cc.previousValue)
	value := atomic.LoadUint64(&cc.value)
	if value == previous {
		return nil
	}
	timestamp := time.Now().UnixNano() / 1000000
	protoDims := make(
		[]*sfxproto.Dimension,
		len(dimensions),
		len(dimensions)+len(cc.Dimensions))
	for i, v := range dimensions {
		protoDims[i] = v
	}
	for k, v := range cc.Dimensions {
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
	prefixedMetric := metricPrefix + cc.Metric
	if value > math.MaxInt64 {
		return nil
	}
	iv := int64(value)
	return &sfxproto.DataPoint{
		Metric:     &prefixedMetric,
		Timestamp:  &timestamp,
		MetricType: &cumulativeCounterType,
		Dimensions: protoDims,
		Value:      &sfxproto.Datum{IntValue: &iv},
	}
}

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
