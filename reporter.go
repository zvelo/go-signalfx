package signalfx

import (
	"fmt"
	"sync"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// DataPointCallback is a functional callback that can be passed to DataPointCallback as a way
// to have the caller calculate and return their own datapoints
type DataPointCallback func(defaultDims sfxproto.Dimensions) *DataPoints

// Reporter is an object that tracks DataPoints and manages a Client. It is the
// recommended way to send data to SignalFX.
type Reporter struct {
	client             *Client
	defaultDimensions  sfxproto.Dimensions
	datapoints         *DataPoints
	buckets            map[*Bucket]interface{}
	preReportCallbacks []func()
	datapointCallbacks []DataPointCallback
	mu                 sync.Mutex
	oneShots           []*sfxproto.DataPoint
}

// NewReporter returns a new Reporter object. Any dimensions supplied will be
// appended to all DataPoints sent to SignalFX. config is copied, so future
// changes to the external config object are not reflected within the reporter.
func NewReporter(config *Config, defaultDimensions sfxproto.Dimensions) *Reporter {
	return &Reporter{
		client:            NewClient(config),
		defaultDimensions: defaultDimensions,
		datapoints:        NewDataPoints(0),
		buckets:           map[*Bucket]interface{}{},
	}
}

func (r *Reporter) lock() {
	r.mu.Lock()
}

func (r *Reporter) unlock() {
	r.mu.Unlock()
}

// NewBucket creates a new Bucket object that is tracked by the Reporter.
// Buckets are goroutine safe.
func (r *Reporter) NewBucket(metric string, dimensions sfxproto.Dimensions) *Bucket {
	ret := NewBucket(metric, dimensions)

	r.lock()
	defer r.unlock()

	r.buckets[ret] = nil
	return ret
}

// NewCumulative returns a new DataPoint object with type CUMULATIVE_COUNTER.
// val can be any type of int, float, string, nil, pointer to those types, or a
// Getter that returns any of those types. Literal pointers are copied by value.
// Getters that return pointer types should not have their value changed, unless
// atomically, when in a Reporter, except within a PreReportCallback, for
// goroutine safety.
func (r *Reporter) NewCumulative(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	dp, _ := NewCumulative(metric, val, r.defaultDimensions.Append(dims))
	r.datapoints.Add(dp)
	return dp
}

// NewGauge returns a new DataPoint object with type GAUGE. val can be any type
// of int, float, string, nil, pointer to those types, or a Getter that returns
// any of those types. Literal pointers are copied by value. Getters that return
// pointer types should not have their value changed, unless atomically, when in
// a Reporter, except within a PreReportCallback, for goroutine safety.
func (r *Reporter) NewGauge(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	dp, _ := NewGauge(metric, val, r.defaultDimensions.Append(dims))
	r.datapoints.Add(dp)
	return dp
}

// NewCounter returns a new DataPoint object with type COUNTER. val can be any
// type of int, float, string, nil, pointer to those types, or a Getter that
// returns any of those types. Literal pointers are copied by value. Getters
// that return pointer types should not have their value changed, unless
// atomically, when in a Reporter, except within a PreReportCallback, for
// goroutine safety.
func (r *Reporter) NewCounter(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	dp, _ := NewCounter(metric, val, r.defaultDimensions.Append(dims))
	r.datapoints.Add(dp)
	return dp
}

// NewInt32 creates a new DataPoint object with type COUNTER whose value is
// bound to an Int32. All methods on Int32 are goroutine safe and it may also be
// modified in atomic operations as an int32.
func (r *Reporter) NewInt32(metric string, dims sfxproto.Dimensions) (*Int32, *DataPoint) {
	ret := Int32(0)
	dp, _ := NewCounter(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewCumulativeInt32 creates a new DataPoint object with type
// CUMULATIVE_COUNTER whose value is bound to an Int32. All methods on Int32 are
// goroutine safe and it may also be modified in atomic operations as an int32.
func (r *Reporter) NewCumulativeInt32(metric string, dims sfxproto.Dimensions) (*Int32, *DataPoint) {
	ret := Int32(0)
	dp, _ := NewCumulative(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewInt64 creates a new DataPoint object with type COUNTER whose value is
// bound to an Int64. All methods on Int64 are goroutine safe and it may also be
// modified in atomic operations as an int64.
func (r *Reporter) NewInt64(metric string, dims sfxproto.Dimensions) (*Int64, *DataPoint) {
	ret := Int64(0)
	dp, _ := NewCounter(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewCumulativeInt64 creates a new DataPoint object with type
// CUMULATIVE_COUNTER whose value is bound to an Int64. All methods on Int64 are
// goroutine safe and it may also be modified in atomic operations as an int64.
func (r *Reporter) NewCumulativeInt64(metric string, dims sfxproto.Dimensions) (*Int64, *DataPoint) {
	ret := Int64(0)
	dp, _ := NewCumulative(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewUint32 creates a new DataPoint object with type COUNTER whose value is
// bound to an Uint32. All methods on Uint32 are goroutine safe and it may also
// be modified in atomic operations as an uint32.
func (r *Reporter) NewUint32(metric string, dims sfxproto.Dimensions) (*Uint32, *DataPoint) {
	ret := Uint32(0)
	dp, _ := NewCounter(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewCumulativeUint32 creates a new DataPoint object with type
// CUMULATIVE_COUNTER whose value is bound to an Uint32. All methods on Uint32
// are goroutine safe and it may also be modified in atomic operations as an
// uint32.
func (r *Reporter) NewCumulativeUint32(metric string, dims sfxproto.Dimensions) (*Uint32, *DataPoint) {
	ret := Uint32(0)
	dp, _ := NewCumulative(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewUint64 creates a new DataPoint object with type COUNTER whose value is
// bound to an Uint64. All methods on Uint64 are goroutine safe and it may also be
// modified in atomic operations as an uint64.
func (r *Reporter) NewUint64(metric string, dims sfxproto.Dimensions) (*Uint64, *DataPoint) {
	ret := Uint64(0)
	dp, _ := NewCounter(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewCumulativeUint64 creates a new DataPoint object with type
// CUMULATIVE_COUNTER whose value is bound to an Uint64. All methods on Uint64 are
// goroutine safe and it may also be modified in atomic operations as an uint64.
func (r *Reporter) NewCumulativeUint64(metric string, dims sfxproto.Dimensions) (*Uint64, *DataPoint) {
	ret := Uint64(0)
	dp, _ := NewCumulative(metric, &ret, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// AddDataPoint provides a way to manually add DataPoint(s) to be tracked by the
// Reporter
func (r *Reporter) AddDataPoint(vals ...*DataPoint) {
	r.datapoints.Add(vals...)
}

// AddDataPoints provides a way to manually add DataPoint to be tracked by the
// Reporter
func (r *Reporter) AddDataPoints(dps *DataPoints) {
	r.datapoints.Append(dps)
}

// RemoveDataPoint takes DataPoint(s)out of the set being tracked by the
// Reporter
func (r *Reporter) RemoveDataPoint(dps ...*DataPoint) {
	r.datapoints.Remove(dps...)
}

// RemoveDataPoints takes DataPoints out of the set being tracked by the
// Reporter
func (r *Reporter) RemoveDataPoints(dps *DataPoints) {
	r.datapoints.RemoveDataPoints(dps)
}

// RemoveBucket takes Bucket(s) out of the set being tracked by the Reporter
func (r *Reporter) RemoveBucket(bs ...*Bucket) {
	r.lock()
	defer r.unlock()

	for _, b := range bs {
		delete(r.buckets, b)
	}
}

// AddPreReportCallback adds a function that is called before Report().  This is useful for refetching
// things like runtime.Memstats() so they are only fetched once per report() call. If a DataPoint
func (r *Reporter) AddPreReportCallback(f func()) {
	r.lock()
	defer r.unlock()
	r.preReportCallbacks = append(r.preReportCallbacks, f)
}

// AddDataPointsCallback adds a callback that itself will generate datapoints to report
func (r *Reporter) AddDataPointsCallback(f DataPointCallback) {
	r.lock()
	defer r.unlock()
	r.datapointCallbacks = append(r.datapointCallbacks, f)
}

// Report sends all tracked DataPoints to SignalFX. PreReportCallbacks will be
// run before building the dataset to send. DataPoint callbacks will be executed
// and added to the dataset, but do not become tracked by the Reporter.
func (r *Reporter) Report(ctx context.Context) (*DataPoints, error) {
	if ctx == nil {
		ctx = context.Background()
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.lock()
	defer r.unlock()

	for _, f := range r.preReportCallbacks {
		f()
	}

	ret := r.datapoints.Clone()

	for _, f := range r.datapointCallbacks {
		ret.Append(f(r.defaultDimensions))
	}

	for b := range r.buckets {
		ret.Append(b.DataPoints(r.defaultDimensions))
	}

	var (
		counters           []*DataPoint
		cumulativeCounters []*DataPoint
	)
	ret = ret.filter(func(dp *DataPoint) bool {
		if err := dp.update(); err != nil {
			return false
		}

		switch *dp.pdp.MetricType {
		case sfxproto.MetricType_COUNTER:
			if dp.pdp.Value.IntValue != nil && *dp.pdp.Value.IntValue != 0 {
				counters = append(counters, dp)
				return true
			}
		case sfxproto.MetricType_CUMULATIVE_COUNTER:
			if !dp.pdp.Equal(dp.previous) {
				cumulativeCounters = append(cumulativeCounters, dp)
				return true
			}
		default:
			return true
		}
		return false
	})
	pdps, err := ret.ProtoDataPoints()
	if err != nil {
		return nil, err
	}

	// append all of the one-shots
	for _, pdp := range r.oneShots {
		pdps.Add(pdp)
	}

	if err = r.client.Submit(ctx, pdps); err != nil {
		return nil, err
	}

	// set submitted counters to zero
	for _, counter := range counters {
		// TODO: what should be done if this fails?
		counter.Set(0)
	}
	// remember submitted cumulative counter values
	for _, counter := range cumulativeCounters {
		counter.previous = counter.pdp
	}

	// and clear the one-shots
	r.oneShots = nil

	return ret, nil
}

func (r *Reporter) Inc(metric string, dimensions map[string]string, delta int64) error {
	r.lock()
	defer r.unlock()

	var protoDims []*sfxproto.Dimension
	for k, v := range dimensions {
		protoDims = append(protoDims, &sfxproto.Dimension{Key: &k, Value: &v})
	}
	timestamp := time.Now().Unix()
	metricType := sfxproto.MetricType_COUNTER
	dp := &sfxproto.DataPoint{
		Metric:     &metric,
		Timestamp:  &timestamp,
		MetricType: &metricType,
		Dimensions: protoDims,
		Value:      &sfxproto.Datum{IntValue: &delta},
	}
	r.oneShots = append(r.oneShots, dp)
	fmt.Println("oneShots", len(r.oneShots))
	return nil
}
