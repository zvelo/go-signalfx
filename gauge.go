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

func (g *Gauge) dataPoint() *dataPoint {
	return &dataPoint{
		Metric:     g.Metric,
		Timestamp:  time.Now(),
		Type:       GaugeType,
		Dimensions: g.Dimensions,
		Value:      atomic.LoadInt64(&g.value),
	}
}
