package signalfx

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// MetricCallback is a functional callback that can be passed to MetricCallback as a way
// to have the caller calculate and return their own datapoints
type MetricCallback func(defaultDims sfxproto.Dimensions) *Metrics

type Reporter struct {
	client             *Client
	defaultDimensions  sfxproto.Dimensions
	metrics            *Metrics
	buckets            map[*Bucket]interface{}
	preReportCallbacks []func()
	metricCallbacks    []MetricCallback
	lock               sync.Mutex
}

func NewReporter(config *Config, defaultDimensions sfxproto.Dimensions) *Reporter {
	return &Reporter{
		client:            NewClient(config),
		defaultDimensions: defaultDimensions,
		metrics:           NewMetrics(0),
		buckets:           map[*Bucket]interface{}{},
	}
}

func (r *Reporter) NewBucket(metricName string, dimensions sfxproto.Dimensions) *Bucket {
	ret := NewBucket(metricName, dimensions)

	r.lock.Lock()
	defer r.lock.Unlock()

	r.buckets[ret] = nil
	return ret
}

func (r *Reporter) NewCumulative(metricName string, val interface{}, dims sfxproto.Dimensions) *Metric {
	m, _ := NewCumulative(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) NewGauge(metricName string, val interface{}, dims sfxproto.Dimensions) *Metric {
	m, _ := NewGauge(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) NewCounter(metricName string, val interface{}, dims sfxproto.Dimensions) *Metric {
	m, _ := NewCounter(metricName, val, r.defaultDimensions.Concat(dims))
	r.metrics.Add(m)
	return m
}

func (r *Reporter) AddMetric(vals ...*Metric) {
	r.metrics.Add(vals...)
}

func (r *Reporter) AddMetrics(ms *Metrics) {
	r.metrics.Concat(ms)
}

func (r *Reporter) RemoveMetric(ms ...*Metric) {
	r.metrics.Remove(ms...)
}

func (r *Reporter) RemoveMetrics(ms *Metrics) {
	r.metrics.RemoveMetrics(ms)
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

// AddMetricsCallback adds a callback that itself will generate datapoints to report
func (r *Reporter) AddMetricsCallback(f MetricCallback) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.metricCallbacks = append(r.metricCallbacks, f)
}

func (r *Reporter) Report(ctx context.Context) (*Metrics, error) {
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

	ret := r.metrics.Clone()

	for _, f := range r.metricCallbacks {
		ret.Concat(f(r.defaultDimensions))
	}

	for b := range r.buckets {
		ret.Concat(b.Metrics(r.defaultDimensions))
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
