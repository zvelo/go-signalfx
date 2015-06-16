package signalfx

import (
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/zvelo/go-signalfx/sfxproto"
)

var (
	// ErrIllegalType is returned when trying to set the Datum value using an
	// unsupported type
	ErrIllegalType = fmt.Errorf("illegal value type")
)

// A DataPoint is a light wrapper around sfxproto.DataPoint. It adds the ability to
// set values via callback by using the Getter Interface. Additionally, all
// operations on it are goroutine/thread safe.
type DataPoint struct {
	dp   *sfxproto.ProtoDataPoint
	get  Getter
	lock sync.Mutex
}

// NewDataPoint creates a new DataPoint. val can be nil, any int type, any float type, a
// string, a pointer to any of those types or a Getter that returns any of those
// types.
func NewDataPoint(metricType sfxproto.MetricType, metric string, val interface{}, dims sfxproto.Dimensions) (*DataPoint, error) {
	ret := &DataPoint{
		dp: &sfxproto.ProtoDataPoint{
			Metric:     proto.String(metric),
			MetricType: metricType.Enum(),
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

// Equal returns whether or not two DataPoint objects are equal
func (m *DataPoint) Equal(val *DataPoint) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	val.lock.Lock()
	defer val.lock.Unlock()

	if !m.dp.Equal(val.dp) {
		return false
	}

	return m.get == val.get
}

// Clone returns a DataPoint with a deep copy of the underlying DataPoint. If there
// is a Getter, the interface is copied, but it is not a deep copy.
func (m *DataPoint) Clone() *DataPoint {
	m.lock.Lock()
	defer m.lock.Unlock()

	return &DataPoint{
		dp:  m.dp.Clone(),
		get: m.get,
	}
}

// Time returns the timestamp of the DataPoint
func (m *DataPoint) Time() time.Time {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.dp.Time()
}

// SetTime sets the timestamp of the DataPoint
func (m *DataPoint) SetTime(t time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dp.SetTime(t)
}

// Metric returns the metric name of the DataPoint
func (m *DataPoint) Metric() string {
	m.lock.Lock()
	defer m.lock.Unlock()

	return *m.dp.Metric
}

// SetMetric sets the metric name of the DataPoint
func (m *DataPoint) SetMetric(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dp.Metric = proto.String(name)
}

// Type returns the MetricType of the DataPoint
func (m *DataPoint) Type() sfxproto.MetricType {
	m.lock.Lock()
	defer m.lock.Unlock()

	return *m.dp.MetricType
}

// Dimensions returns a copy of the dimensions of the DataPoint. Changes are not
// reflected inside the DataPoint itself.
func (m *DataPoint) Dimensions() sfxproto.Dimensions {
	m.lock.Lock()
	defer m.lock.Unlock()

	return sfxproto.NewDimensions(m.dp.Dimensions)
}

// SetDimension adds or overwrites the dimension at key with value
func (m *DataPoint) SetDimension(key, value string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, dim := range m.dp.Dimensions {
		if *dim.Key == key {
			*dim.Value = value
			return
		}
	}

	m.dp.Dimensions = append(m.dp.Dimensions, &sfxproto.Dimension{
		Key:   proto.String(key),
		Value: proto.String(value),
	})
}

// SetDimensions adds or overwrites multiple dimensions
func (m *DataPoint) SetDimensions(dims sfxproto.Dimensions) {
	for key, value := range dims {
		m.SetDimension(key, value)
	}
}

// RemoveDimension removes one or more dimensions with the given keys
func (m *DataPoint) RemoveDimension(keys ...string) {
	m.lock.Lock()
	defer m.lock.Unlock()

Loop:
	for _, key := range keys {
		for i, dim := range m.dp.Dimensions {
			if *dim.Key == key {
				m.dp.Dimensions = append(m.dp.Dimensions[:i], m.dp.Dimensions[i+1:]...)
				continue Loop
			}
		}
	}
}

// StrValue returns the string value of the Datum of the underlying DataPoint
func (m *DataPoint) StrValue() string {
	m.update() // ignore error as it is reflected in the returned value

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.dp.Value.StrValue == nil {
		return ""
	}

	return *m.dp.Value.StrValue
}

// IntValue returns the integer value of the Datum of the underlying DataPoint
func (m *DataPoint) IntValue() int64 {
	m.update() // ignore error as it is reflected in the returned value

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.dp.Value.IntValue == nil {
		return 0
	}

	return *m.dp.Value.IntValue
}

// DoubleValue returns the integer value of the Datum of the underlying DataPoint
func (m *DataPoint) DoubleValue() float64 {
	m.update() // ignore error as it is reflected in the returned value

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.dp.Value.DoubleValue == nil {
		return 0
	}

	return *m.dp.Value.DoubleValue
}

// Set the value of the DataPoint. It can be nil, any int type, any float type, a
// string, a pointer to any of those types or a Getter that returns any of those
// types.
func (m *DataPoint) Set(val interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dp.Value.Reset()
	m.get = nil

	if val == nil {
		return nil
	}

	if get, ok := val.(Getter); ok {
		m.get = get

		var err error
		if val, err = m.get.Get(); err != nil {
			return err
		}
	}

	if val, err := toInt64(val); err == nil {
		m.dp.Value.IntValue = proto.Int64(val)
		return nil
	}

	if val, err := toFloat64(val); err == nil {
		m.dp.Value.DoubleValue = proto.Float64(val)
		return nil
	}

	if val, err := toString(val); err == nil {
		m.dp.Value.StrValue = proto.String(val)
		return nil
	}

	return ErrIllegalType
}

// NewCumulative returns a new DataPoint set to type CUMULATIVE_COUNTER
func NewCumulative(metric string, val interface{}, dims sfxproto.Dimensions) (*DataPoint, error) {
	return NewDataPoint(sfxproto.MetricType_CUMULATIVE_COUNTER, metric, val, dims)
}

// NewGauge returns a new DataPoint set to type GAUGE
func NewGauge(metric string, val interface{}, dims sfxproto.Dimensions) (*DataPoint, error) {
	return NewDataPoint(sfxproto.MetricType_GAUGE, metric, val, dims)
}

// NewCounter returns a new DataPoint set to type COUNTER
func NewCounter(metric string, val interface{}, dims sfxproto.Dimensions) (*DataPoint, error) {
	return NewDataPoint(sfxproto.MetricType_COUNTER, metric, val, dims)
}

func (m *DataPoint) update() error {
	m.lock.Lock()

	if m.get == nil {
		m.lock.Unlock()
		return nil
	}

	m.lock.Unlock()
	return m.Set(m.get)
}
