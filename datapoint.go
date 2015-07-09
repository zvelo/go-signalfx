package signalfx

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/zvelo/go-signalfx/sfxproto"
)

var (
	// ErrIllegalType is returned when trying to set the Datum value using an
	// unsupported type
	ErrIllegalType = fmt.Errorf("illegal value type")

	// ErrNoMetricName is returned when trying to create a DataPoint without a
	// Metric name
	ErrNoMetricName = fmt.Errorf("no metric name")
)

// A DataPoint is a light wrapper around sfxproto.DataPoint. It adds the
// ability to set values via callback by using the Getter Interface.
// Additionally, all operations on it are goroutine/thread safe.
type DataPoint struct {
	pdp *sfxproto.DataPoint
	get Getter
	mu  sync.Mutex
}

// NewDataPoint creates a new DataPoint. val can be nil, any int type, any float
// type, a string, a pointer to any of those types or a Getter that returns any
// of those types.
func NewDataPoint(metricType sfxproto.MetricType, metric string, val interface{}, dims sfxproto.Dimensions) (*DataPoint, error) {
	if len(metric) == 0 {
		return nil, ErrNoMetricName
	}

	ret := &DataPoint{
		pdp: &sfxproto.DataPoint{
			Metric:     proto.String(metric),
			MetricType: metricType.Enum(),
			Value:      &sfxproto.Datum{},
		},
	}

	if dims != nil {
		ret.pdp.Dimensions = dims.List()
	}

	ret.SetTime(time.Now())

	if err := ret.Set(val); err != nil {
		return nil, err
	}

	return ret, nil
}

// Equal returns whether or not two DataPoint objects are equal
func (dp *DataPoint) Equal(val *DataPoint) bool {
	dp.lock()
	defer dp.unlock()

	val.mu.Lock() // FIXME: why dp.lock() above but val.mu.Lock() here?
	defer val.mu.Unlock()

	if dp.get != nil || val.get != nil {
		return dp.get == val.get
	}

	return dp.pdp.Equal(val.pdp)
}

// Clone returns a DataPoint with a deep copy of the underlying DataPoint. If there
// is a Getter, the interface is copied, but it is not a deep copy.
func (dp *DataPoint) Clone() *DataPoint {
	dp.lock()
	defer dp.unlock()

	return &DataPoint{
		pdp: dp.pdp.Clone(),
		get: dp.get,
	}
}

// Time returns the timestamp of the DataPoint
func (dp *DataPoint) Time() time.Time {
	dp.lock()
	defer dp.unlock()

	return dp.pdp.Time()
}

// SetTime sets the timestamp of the DataPoint
func (dp *DataPoint) SetTime(t time.Time) {
	dp.lock()
	defer dp.unlock()

	dp.pdp.SetTime(t)
}

// Metric returns the metric name of the DataPoint
func (dp *DataPoint) Metric() string {
	// TODO: is there any need to mutex this, since strings are
	// immutable and thus it will never be corrupted?
	dp.lock()
	defer dp.unlock()

	return *dp.pdp.Metric
}

// SetMetric sets the metric name of the DataPoint
func (dp *DataPoint) SetMetric(name string) {
	dp.lock()
	defer dp.unlock()

	dp.pdp.Metric = proto.String(name)
}

// Type returns the MetricType of the DataPoint
func (dp *DataPoint) Type() sfxproto.MetricType {
	dp.lock()
	defer dp.unlock()

	return *dp.pdp.MetricType
}

// Dimensions returns a copy of the dimensions of the DataPoint. Changes are not
// reflected inside the DataPoint itself.
func (dp *DataPoint) Dimensions() sfxproto.Dimensions {
	dp.lock()
	defer dp.unlock()

	return sfxproto.NewDimensions(dp.pdp.Dimensions)
}

// SetDimension adds or overwrites the dimension at key with value. If the key
// or value is empty, no changes are made
func (dp *DataPoint) SetDimension(key, value string) {
	if key == "" || value == "" {
		return
	}

	dp.lock()
	defer dp.unlock()

	for _, dim := range dp.pdp.Dimensions {
		if *dim.Key == key {
			*dim.Value = value
			return
		}
	}

	dp.pdp.Dimensions = append(dp.pdp.Dimensions, &sfxproto.Dimension{
		Key:   proto.String(key),
		Value: proto.String(value),
	})
}

// SetDimensions adds or overwrites multiple dimensions
func (dp *DataPoint) SetDimensions(dims sfxproto.Dimensions) {
	for key, value := range dims {
		dp.SetDimension(key, value)
	}
}

// RemoveDimension removes one or more dimensions with the given keys
func (dp *DataPoint) RemoveDimension(keys ...string) {
	dp.lock()
	defer dp.unlock()

Loop:
	for _, key := range keys {
		for i, dim := range dp.pdp.Dimensions {
			if *dim.Key == key {
				dp.pdp.Dimensions = append(dp.pdp.Dimensions[:i], dp.pdp.Dimensions[i+1:]...)
				continue Loop
			}
		}
	}
}

// StrValue returns the string value of the Datum of the underlying DataPoint
func (dp *DataPoint) StrValue() string {
	dp.update() // ignore error as it is reflected in the returned value

	dp.lock()
	defer dp.unlock()

	if dp.pdp.Value.StrValue == nil {
		return ""
	}

	return *dp.pdp.Value.StrValue
}

// IntValue returns the integer value of the Datum of the underlying DataPoint
func (dp *DataPoint) IntValue() int64 {
	dp.update() // ignore error as it is reflected in the returned value

	dp.lock()
	defer dp.unlock()

	if dp.pdp.Value.IntValue == nil {
		return 0
	}

	return *dp.pdp.Value.IntValue
}

// DoubleValue returns the integer value of the Datum of the underlying DataPoint
func (dp *DataPoint) DoubleValue() float64 {
	dp.update() // ignore error as it is reflected in the returned value

	dp.lock()
	defer dp.unlock()

	if dp.pdp.Value.DoubleValue == nil {
		return 0
	}

	return *dp.pdp.Value.DoubleValue
}

// Set the value of the DataPoint. It can be nil, any int type, any float type, a
// string, a pointer to any of those types or a Getter that returns any of those
// types.
func (dp *DataPoint) Set(val interface{}) error {
	dp.lock()
	defer dp.unlock()

	dp.pdp.Value.Reset()
	dp.get = nil

	if val == nil {
		return nil
	}

	if get, ok := val.(Getter); ok {
		dp.get = get

		var err error
		if val, err = dp.get.Get(); err != nil {
			return err
		}
	}

	if val, err := toInt64(val); err == nil {
		dp.pdp.Value.IntValue = proto.Int64(val)
		return nil
	}

	if val, err := toFloat64(val); err == nil {
		dp.pdp.Value.DoubleValue = proto.Float64(val)
		return nil
	}

	if val, err := toString(val); err == nil {
		dp.pdp.Value.StrValue = proto.String(val)
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

func (dp *DataPoint) lock() {
	dp.mu.Lock()
}

func (dp *DataPoint) unlock() {
	dp.mu.Unlock()
}

func (dp *DataPoint) update() error {
	dp.lock()

	if dp.get == nil {
		dp.unlock()
		return nil
	}

	dp.unlock()
	return dp.Set(dp.get)
}
