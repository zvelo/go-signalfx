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
	lock               sync.Mutex
}

func NewReporter(config *Config, defaultDimensions sfxproto.Dimensions) *Reporter {
	return &Reporter{
		client:            NewClient(config),
		defaultDimensions: defaultDimensions,
		datapoints:        NewDataPoints(0),
		buckets:           map[*Bucket]interface{}{},
	}
}

func (r *Reporter) NewBucket(metric string, dimensions sfxproto.Dimensions) *Bucket {
	ret := NewBucket(metric, dimensions)

	r.lock.Lock()
	defer r.lock.Unlock()

	r.buckets[ret] = nil
	return ret
}

func (r *Reporter) NewCumulative(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	m, _ := NewCumulative(metric, val, r.defaultDimensions.Concat(dims))
	r.datapoints.Add(m)
	return m
}

func (r *Reporter) NewGauge(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	m, _ := NewGauge(metric, val, r.defaultDimensions.Concat(dims))
	r.datapoints.Add(m)
	return m
}

func (r *Reporter) NewCounter(metric string, val interface{}, dims sfxproto.Dimensions) *DataPoint {
	m, _ := NewCounter(metric, val, r.defaultDimensions.Concat(dims))
	r.datapoints.Add(m)
	return m
}

func (r *Reporter) NewIncrementer(metric string, dims sfxproto.Dimensions) *Incrementer {
	inc := NewIncrementer(0)
	m, _ := NewCounter(metric, inc, r.defaultDimensions.Concat(dims))
	if m == nil {
		return nil
	}
	r.datapoints.Add(m)
	return inc
}

func (r *Reporter) NewCumulativeIncrementer(metric string, dims sfxproto.Dimensions) *Incrementer {
	inc := NewIncrementer(0)
	m, _ := NewCumulative(metric, inc, r.defaultDimensions.Concat(dims))
	if m == nil {
		return nil
	}
	r.datapoints.Add(m)
	return inc
}

func (r *Reporter) AddDataPoint(vals ...*DataPoint) {
	r.datapoints.Add(vals...)
}

func (r *Reporter) AddDataPoints(ms *DataPoints) {
	r.datapoints.Concat(ms)
}

func (r *Reporter) RemoveDataPoint(ms ...*DataPoint) {
	r.datapoints.Remove(ms...)
}

func (r *Reporter) RemoveDataPoints(ms *DataPoints) {
	r.datapoints.RemoveDataPoints(ms)
}

func (r *Reporter) RemoveBucket(bs ...*Bucket) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, b := range bs {
		delete(r.buckets, b)
	}
}

// AddPreReportCallback adds a function that is called before Report().  This is useful for refetching
// things like runtime.Memstats() so they are only fetched once per report() call
func (r *Reporter) AddPreReportCallback(f func()) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.preReportCallbacks = append(r.preReportCallbacks, f)
}

// AddDataPointsCallback adds a callback that itself will generate datapoints to report
func (r *Reporter) AddDataPointsCallback(f DataPointCallback) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.datapointCallbacks = append(r.datapointCallbacks, f)
}

func (r *Reporter) Report(ctx context.Context) (*DataPoints, error) {
	if ctx == nil {
		ctx = context.Background()
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.lock.Lock()
	defer r.lock.Unlock()

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

	dps, err := ret.DataPoints()
	if err != nil {
		return nil, err
	}

	if err = r.client.Submit(ctx, dps); err != nil {
		return nil, err
	}

	return ret, nil
}
