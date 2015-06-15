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

// A Metric is a light wrapper around sfxproto.DataPoint. It adds the ability to
// set values asynchronously by using Getter values. Additionally, all
// operations on it are goroutine/thread safe.
type Metric struct {
	dp   *sfxproto.DataPoint
	get  Getter
	lock sync.Mutex
}

// New creates a new Metric. val can be nil, any int type, any float type, a
// string or a Getter that returns any of those types.
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

// Equal returns whether or not two Metric objects are equal
func (m *Metric) Equal(val *Metric) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	val.lock.Lock()
	defer val.lock.Unlock()

	if !m.dp.Equal(val.dp) {
		return false
	}

	return m.get == val.get
}

// Clone returns a Metric with a deep copy of the underlying DataPoint. If there
// is a Getter, the interface is copied, but it is not a deep copy.
func (m *Metric) Clone() *Metric {
	m.lock.Lock()
	defer m.lock.Unlock()

	return &Metric{
		dp:  m.dp.Clone(),
		get: m.get,
	}
}

// Time returns the timestamp of the Metric
func (m *Metric) Time() time.Time {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Time()
}

// SetTime sets the timestamp of the Metric
func (m *Metric) SetTime(t time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dp.SetTime(t)
}

// Name returns the name of the Metric
func (m *Metric) Name() string {
	// read-only, no need to lock
	return m.dp.Metric
}

// Type returns the MetricType of the Metric
func (m *Metric) Type() sfxproto.MetricType {
	// read-only, no need to lock
	return m.dp.MetricType
}

// Dimensions returns a copy of the dimensions of the Metric. Changes are not
// reflected inside the Metric itself.
func (m *Metric) Dimensions() sfxproto.Dimensions {
	// read-only, no need to lock
	return sfxproto.NewDimensions(m.dp.Dimensions)
}

// StrValue returns the string value of the Datum of the underlying DataPoint
func (m *Metric) StrValue() string {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Value.StrValue
}

// IntValue returns the integer value of the Datum of the underlying DataPoint
func (m *Metric) IntValue() int64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Value.IntValue
}

// DoubleValue returns the integer value of the Datum of the underlying DataPoint
func (m *Metric) DoubleValue() float64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Value.DoubleValue
}

// Set the value of the Metric. It can be nil, any int type, any float type, a
// string or a Getter that returns any of those types.
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

// NewCumulative returns a new Metric set to type CUMULATIVE_COUNTER
func NewCumulative(metric string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	return New(sfxproto.MetricType_CUMULATIVE_COUNTER, metric, val, dims)
}

// NewGauge returns a new Metric set to type GAUGE
func NewGauge(metric string, val interface{}, dims sfxproto.Dimensions) (*Metric, error) {
	return New(sfxproto.MetricType_GAUGE, metric, val, dims)
}

// NewCounter returns a new Metric set to type COUNTER
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
