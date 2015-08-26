package signalfx

import (
	"sync/atomic"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

type Gauge struct {
	Metric     string
	Dimensions map[string]string
	value      int64
}

var gaugeType = sfxproto.MetricType_GAUGE

func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

func (g *Gauge) dataPoint(
	metricPrefix string,
	dimensions []*sfxproto.Dimension,
) *sfxproto.DataPoint {
	timestamp := time.Now().UnixNano() / 1000000
	protoDims := make(
		[]*sfxproto.Dimension,
		len(dimensions),
		len(dimensions)+len(g.Dimensions))
	for i, v := range dimensions {
		protoDims[i] = v
	}
	for k, v := range g.Dimensions {
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
	// have to create a new addressable value, since metric is
	// stored as a pointer…
	prefixedMetric := metricPrefix + g.Metric
	value := atomic.LoadInt64(&g.value)
	return &sfxproto.DataPoint{
		Metric:     &prefixedMetric,
		Timestamp:  &timestamp,
		MetricType: &gaugeType,
		Dimensions: protoDims,
		Value:      &sfxproto.Datum{IntValue: &value},
	}
}
