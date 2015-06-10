package sfxreporter

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxclient"
	"github.com/zvelo/go-signalfx/sfxconfig"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// DataPointCallback is a functional callback that can be passed to DataPointCallback as a way
// to have the caller calculate and return their own datapoints
type DataPointCallback func(defaultDims sfxproto.Dimensions) *sfxproto.DataPoints

type Reporter struct {
	client              *sfxclient.Client
	defaultDimensions   sfxproto.Dimensions
	metrics             *Metrics
	buckets             map[*Bucket]interface{}
	preCollectCallbacks []func()
	dataPointCallbacks  []DataPointCallback
	lock                sync.Mutex
}

func New(config *sfxconfig.Config, defaultDimensions sfxproto.Dimensions) *Reporter {
	return &Reporter{
		client:            sfxclient.New(config),
		defaultDimensions: defaultDimensions,
		metrics:           NewMetrics(0),
	}
}

func (r *Reporter) Bucket(metricName string, dimensions sfxproto.Dimensions) *Bucket {
	ret := NewBucket(metricName, dimensions)

	r.lock.Lock()
	defer r.lock.Unlock()

	r.buckets[ret] = nil
	return ret
}

func (r *Reporter) Cumulative(metricName string, val interface{}, dims sfxproto.Dimensions) *Metric {
	m, _ := NewCumulative(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) Gauge(metricName string, val interface{}, dims sfxproto.Dimensions) *Metric {
	m, _ := NewGauge(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) Counter(metricName string, val interface{}, dims sfxproto.Dimensions) *Metric {
	m, _ := NewCounter(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) RemoveMetric(ms ...*Metric) {
	r.metrics.Remove(ms...)
}

func (r *Reporter) RemoveMetrics(ms *Metrics) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for m := range ms.metrics {
		r.RemoveMetric(m)
	}
}

func (r *Reporter) RemoveBucket(bs ...*Bucket) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, b := range bs {
		delete(r.buckets, b)
	}
}

// AddPreCollectCallback adds a function that is called before Report().  This is useful for refetching
// things like runtime.Memstats() so they are only fetched once per report() call
func (r *Reporter) AddPreCollectCallback(f func()) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.preCollectCallbacks = append(r.preCollectCallbacks, f)
}

// AddDataPointCallback adds a callback that itself will generate datapoints to report
func (r *Reporter) AddDataPointCallback(f DataPointCallback) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.dataPointCallbacks = append(r.dataPointCallbacks, f)
}

func (r *Reporter) Report(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, f := range r.preCollectCallbacks {
		f()
	}

	dps, err := r.metrics.DataPoints()
	if err != nil {
		return err
	}

	for _, f := range r.dataPointCallbacks {
		dps = dps.Concat(f(r.defaultDimensions))
	}

	for b := range r.buckets {
		dps = dps.Concat(b.DataPoints(r.defaultDimensions))
	}

	return r.client.Submit(ctx, dps)
}
