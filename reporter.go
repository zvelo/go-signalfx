package sfxreporter

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxclient"
	"github.com/zvelo/go-signalfx/sfxmetric"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// DataPointCallback is a functional callback that can be passed to DataPointCallback as a way
// to have the caller calculate and return their own datapoints
type DataPointCallback func(defaultDims sfxproto.Dimensions) *sfxproto.DataPoints

type Reporter struct {
	client             *sfxclient.Client
	defaultDimensions  sfxproto.Dimensions
	metrics            *sfxmetric.Metrics
	buckets            map[*sfxmetric.Bucket]interface{}
	preReportCallbacks []func()
	dataPointCallbacks []DataPointCallback
	lock               sync.Mutex
}

func NewReporter(config *sfxclient.Config, defaultDimensions sfxproto.Dimensions) *Reporter {
	return &Reporter{
		client:            sfxclient.NewClient(config),
		defaultDimensions: defaultDimensions,
		metrics:           sfxmetric.NewMetrics(0),
	}
}

func (r *Reporter) Bucket(metricName string, dimensions sfxproto.Dimensions) *sfxmetric.Bucket {
	ret := sfxmetric.NewBucket(metricName, dimensions)

	r.lock.Lock()
	defer r.lock.Unlock()

	r.buckets[ret] = nil
	return ret
}

func (r *Reporter) Cumulative(metricName string, val interface{}, dims sfxproto.Dimensions) *sfxmetric.Metric {
	m, _ := sfxmetric.NewCumulative(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) Gauge(metricName string, val interface{}, dims sfxproto.Dimensions) *sfxmetric.Metric {
	m, _ := sfxmetric.NewGauge(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) Counter(metricName string, val interface{}, dims sfxproto.Dimensions) *sfxmetric.Metric {
	m, _ := sfxmetric.NewCounter(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) RemoveMetric(ms ...*sfxmetric.Metric) {
	r.metrics.Remove(ms...)
}

func (r *Reporter) RemoveMetrics(ms *sfxmetric.Metrics) {
	r.metrics.RemoveMetrics(ms)
}

func (r *Reporter) RemoveBucket(bs ...*sfxmetric.Bucket) {
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

// AddDataPointCallback adds a callback that itself will generate datapoints to report
func (r *Reporter) AddDataPointCallback(f DataPointCallback) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.dataPointCallbacks = append(r.dataPointCallbacks, f)
}

func (r *Reporter) Report(ctx context.Context) (*sfxproto.DataPoints, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, f := range r.preReportCallbacks {
		f()
	}

	dps, err := r.metrics.DataPoints()
	if err != nil {
		return nil, err
	}

	for _, f := range r.dataPointCallbacks {
		dps = dps.Concat(f(r.defaultDimensions))
	}

	for b := range r.buckets {
		tmp, err := b.Metrics(r.defaultDimensions).DataPoints()
		if err != nil {
			return nil, err
		}
		dps = dps.Concat(tmp)
	}

	if err = r.client.Submit(ctx, dps); err != nil {
		return nil, err
	}

	return dps, nil
}
