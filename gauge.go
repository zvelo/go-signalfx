package signalfx

import (
	"fmt"
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

type WrappedGauge struct {
	Metric     string
	Dimensions map[string]string
	Value      Getter
}

func (c *WrappedGauge) dataPoint() *dataPoint {
	gottenValue, err := c.Value.Get()
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	value, err := toInt64(gottenValue)
	if err != nil {
		fmt.Println("error2:", err)
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
