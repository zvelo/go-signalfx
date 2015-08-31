package signalfx

import (
	"sync/atomic"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

type Gauge struct {
	metric     string
	dimensions map[string]string
	value      int64
}

var gaugeType = sfxproto.MetricType_GAUGE

func NewGauge(metric string, dimensions map[string]string, value int64) *Gauge {
	return &Gauge{metric: metric, dimensions: dimensions, value: value}
}

func (g *Gauge) Record(value int64) {
	atomic.StoreInt64(&g.value, value)
}

func (g *Gauge) DataPoint() *DataPoint {
	return &DataPoint{
		Metric:     g.metric,
		Timestamp:  time.Now(),
		Type:       GaugeType,
		Dimensions: g.dimensions,
		Value:      atomic.LoadInt64(&g.value),
	}
}

type WrappedGauge struct {
	metric     string
	dimensions map[string]string
	value      Getter
}

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
