package signalfx

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// DataPointCallback is a functional callback that can be passed to DataPointCallback as a way
// to have the caller calculate and return their own datapoints
type DataPointCallback func(defaultDims sfxproto.Dimensions) *DataPoints

type Reporter struct {
	client             *Client
	defaultDimensions  sfxproto.Dimensions
	datapoints         *DataPoints
	buckets            map[*Bucket]interface{}
	preReportCallbacks []func()
	datapointCallbacks []DataPointCallback
	mu                 sync.Mutex
}

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

func (r *Reporter) NewBucket(metric string, dimensions sfxproto.Dimensions) *Bucket {
	ret := NewBucket(metric, dimensions)

	r.lock()
	defer r.unlock()

	r.buckets[ret] = nil
	return ret
}

func (r *Reporter) NewCumulative(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	dp, _ := NewCumulative(metric, val, r.defaultDimensions.Concat(dims))
	r.datapoints.Add(dp)
	return dp
}

func (r *Reporter) NewGauge(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	dp, _ := NewGauge(metric, val, r.defaultDimensions.Concat(dims))
	r.datapoints.Add(dp)
	return dp
}

func (r *Reporter) NewCounter(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	dp, _ := NewCounter(metric, val, r.defaultDimensions.Concat(dims))
	r.datapoints.Add(dp)
	return dp
}

func (r *Reporter) NewInc(metric string, dims sfxproto.Dimensions) *Inc {
	inc := NewInc(0)
	dp, _ := NewCounter(metric, inc, r.defaultDimensions.Concat(dims))
	if dp == nil {
		return nil
	}
	r.datapoints.Add(dp)
	return inc
}

func (r *Reporter) NewCumulativeInc(metric string, dims sfxproto.Dimensions) *Inc {
	inc := NewInc(0)
	dp, _ := NewCumulative(metric, inc, r.defaultDimensions.Concat(dims))
	if dp == nil {
		return nil
	}
	r.datapoints.Add(dp)
	return inc
}

func (r *Reporter) AddDataPoint(vals ...*DataPoint) {
	r.datapoints.Add(vals...)
}

func (r *Reporter) AddDataPoints(dps *DataPoints) {
	r.datapoints.Concat(dps)
}

func (r *Reporter) RemoveDataPoint(dps ...*DataPoint) {
	r.datapoints.Remove(dps...)
}

func (r *Reporter) RemoveDataPoints(dps *DataPoints) {
	r.datapoints.RemoveDataPoints(dps)
}

func (r *Reporter) RemoveBucket(bs ...*Bucket) {
	r.lock()
	defer r.unlock()

	for _, b := range bs {
		delete(r.buckets, b)
	}
}

// AddPreReportCallback adds a function that is called before Report().  This is useful for refetching
// things like runtime.Memstats() so they are only fetched once per report() call
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
		ret.Concat(f(r.defaultDimensions))
	}

	for b := range r.buckets {
		ret.Concat(b.DataPoints(r.defaultDimensions))
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
