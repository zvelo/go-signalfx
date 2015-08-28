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

func (g *Gauge) Record(value int64) {
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

type WrappedGauge struct {
	Metric     string
	Dimensions map[string]string
	Value      Getter
}

func WrapGauge(
	metric string,
	dimensions map[string]string,
	value Getter,
) *WrappedGauge {
	return &WrappedGauge{
		Metric:     metric,
		Dimensions: dimensions,
		Value:      value,
	}
}

func (c *WrappedGauge) dataPoint() *dataPoint {
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
		Type:       GaugeType,
		Dimensions: c.Dimensions,
		Value:      value,
	}
}
