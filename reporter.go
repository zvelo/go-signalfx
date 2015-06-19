package signalfx

import (
	"sync"

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

// NewInc creates a new DataPoint object with type COUNTER whose value is bound
// to an incrementer. All operations on the Inc object are goroutine safe.
func (r *Reporter) NewInc(metric string, dims sfxproto.Dimensions) (*Inc, *DataPoint) {
	inc := NewInc(0)
	dp, _ := NewCounter(metric, inc, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return inc, dp
}

// NewCumulativeInc creates a new DataPoint object with type CUMULATIVE_COUNTER
// whose value is bound to an incrementer. All operations on the Inc object are
// goroutine safe.
func (r *Reporter) NewCumulativeInc(metric string, dims sfxproto.Dimensions) (*Inc, *DataPoint) {
	inc := NewInc(0)
	dp, _ := NewCumulative(metric, inc, r.defaultDimensions.Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return inc, dp
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

	pdps, err := ret.ProtoDataPoints()
	if err != nil {
		return nil, err
	}

	if err = r.client.Submit(ctx, pdps); err != nil {
		return nil, err
	}

	return ret, nil
}
