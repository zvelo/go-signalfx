package sfxmetric

import (
	"fmt"
	"sync"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

var (
	// ErrIllegalType is returned when trying to set the Datum value using an
	// unsupported type
	ErrIllegalType = fmt.Errorf("illegal value type")
)

type Metric struct {
	dp   *sfxproto.DataPoint
	get  Getter
	lock sync.Mutex
}

func New(metricType sfxproto.MetricType, metric string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	ret := &Metric{
		dp: &sfxproto.DataPoint{
			Metric:     metric,
			MetricType: metricType,
			Value:      &sfxproto.Datum{},
		},
	}

	if dims != nil {
		ret.dp.Dimensions = dims.List()
	}

	ret.SetTime(time.Now())

	if err := ret.Set(val); err != nil {
		return nil, err
	}

	return ret, nil
}

func (m *Metric) Equals(val *Metric) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	val.lock.Lock()
	defer val.lock.Unlock()

	if !m.dp.Equals(val.dp) {
		return false
	}

	return m.get == val.get
}

func (m *Metric) Clone() *Metric {
	m.lock.Lock()
	defer m.lock.Unlock()

	return &Metric{
		dp:  m.dp.Clone(),
		get: m.get,
	}
}

func (m *Metric) Time() time.Time {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Time()
}

func (m *Metric) SetTime(t time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dp.SetTime(t)
}

func (m *Metric) Metric() string {
	// read-only, no need to lock
	return m.dp.Metric
}

func (m *Metric) MetricType() sfxproto.MetricType {
	// read-only, no need to lock
	return m.dp.MetricType
}

func (m *Metric) Dimensions() sfxproto.Dimensions {
	// read-only, no need to lock
	return sfxproto.NewDimensions(m.dp.Dimensions)
}

func (m *Metric) StrValue() string {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Value.StrValue
}

func (m *Metric) IntValue() int64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Value.IntValue
}

func (m *Metric) DoubleValue() float64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Value.DoubleValue
}

func (m *Metric) Set(val interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dp.Value.Reset()

	if val == nil {
		return nil
	}

	var err error

	if get, ok := val.(Getter); ok {
		m.get = get
		val, err = m.get.Get()
		if err != nil {
			return err
		}
	}

	if m.dp.Value.IntValue, err = toInt64(val); err == nil {
		return nil
	}

	if m.dp.Value.DoubleValue, err = toFloat64(val); err == nil {
		return nil
	}

	if m.dp.Value.StrValue, err = toString(val); err == nil {
		return nil
	}

	return ErrIllegalType
}

func NewCumulative(metric string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	return New(sfxproto.MetricType_CUMULATIVE_COUNTER, metric, val, dims)
}

func NewGauge(metric string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	return New(sfxproto.MetricType_GAUGE, metric, val, dims)
}

func NewCounter(metric string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	return New(sfxproto.MetricType_COUNTER, metric, val, dims)
}

func (m *Metric) update() error {
	m.lock.Lock()

	if m.get == nil {
		m.lock.Unlock()
		return nil
	}

	m.lock.Unlock()
	return m.Set(m.get)
}
