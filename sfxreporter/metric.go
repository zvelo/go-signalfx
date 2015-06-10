package sfxreporter

import (
	"sync"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

type Getter interface {
	Get() interface{}
}

type GetterFunc func() interface{}

func (f GetterFunc) Get() interface{} {
	return f()
}

// ValueGetter is a convenience function for making any value satisfy the Getter
// interface
func ValueGetter(value interface{}) Getter {
	return valueGetter{value}
}

type valueGetter struct {
	value interface{}
}

func (v valueGetter) Get() interface{} {
	return v.value
}

type Metric struct {
	dp  *sfxproto.DataPoint
	get Getter
}

func NewMetric(dp *sfxproto.DataPoint, val interface{}) (*Metric, error) {
	ret := &Metric{
		dp: dp,
	}

	if get, ok := val.(Getter); ok {
		ret.get = get
	} else {
		if err := ret.dp.SetValue(val); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func NewCumulative(metricName string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	dp, err := sfxproto.NewCumulative(metricName, nil, dims)
	if err != nil {
		return nil, err
	}

	return NewMetric(dp, val)
}

func NewGauge(metricName string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	dp, err := sfxproto.NewGauge(metricName, nil, dims)
	if err != nil {
		return nil, err
	}

	return NewMetric(dp, val)
}

func NewCounter(metricName string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	dp, err := sfxproto.NewCounter(metricName, nil, dims)
	if err != nil {
		return nil, err
	}

	return NewMetric(dp, val)
}

func (m *Metric) SetTime(t time.Time) {
	m.dp.SetTime(t)
}

func (m *Metric) Time() time.Time {
	return m.dp.Time()
}

func (m *Metric) SetValue(val interface{}) error {
	return m.dp.SetValue(val)
}

func (m *Metric) Value() *sfxproto.Datum {
	return m.dp.Value
}

func (m *Metric) StrValue() string {
	return m.dp.Value.StrValue
}

func (m *Metric) DoubleValue() float64 {
	return m.dp.Value.DoubleValue
}

func (m *Metric) IntValue() int64 {
	return m.dp.Value.IntValue
}

func (m *Metric) String() string {
	return m.dp.String()
}

type Metrics struct {
	metrics map[*Metric]interface{}
	lock    sync.Mutex
}

func NewMetrics(l int) *Metrics {
	return &Metrics{
		metrics: make(map[*Metric]interface{}, l),
	}
}

func (ms *Metrics) Add(m *Metric) *Metrics {
	if m != nil {
		ms.lock.Lock()
		defer ms.lock.Unlock()

		ms.metrics[m] = nil
	}

	return ms
}

func (ms *Metrics) Remove(vals ...*Metric) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for _, m := range vals {
		delete(ms.metrics, m)
	}
}

func (ms *Metrics) DataPoints() (*sfxproto.DataPoints, error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	ret := sfxproto.NewDataPoints(len(ms.metrics))

	for m := range ms.metrics {
		if m.get != nil {
			if err := m.dp.SetValue(m.get.Get()); err != nil {
				return nil, err
			}
		}

		ret.Add(m.dp)
	}

	return ret, nil
}
