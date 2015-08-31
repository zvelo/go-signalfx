package signalfx

import (
	"math"
	"sync"
	"time"

	"log"

	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// Metrics represent producers of datapoints (i.e., metric time
// series).  While individual datapoints are not goroutine-safe (they
// are owned by one goroutine at a time), metrics should be.
type Metric interface {
	// DataPoint returns the value of the metric at the current
	// point in time.  If it has no value at the present point in
	// time, return nil.
	DataPoint() *DataPoint
}

type HookedMetric interface {
	Metric
	PostReportHook(reportedValue int64)
}

// DataPointCallback is a functional callback that can be passed to
// DataPointCallback as a way to have the caller calculate and return
// their own datapoints
type DataPointCallback func() []DataPoint

// Reporter is an object that tracks DataPoints and manages a Client. It is the
// recommended way to send data to SignalFX.
type Reporter struct {
	client            *Client
	defaultDimensions map[string]string
	//datapoints         *DataPoints
	metrics            map[Metric]struct{}
	buckets            map[*Bucket]interface{}
	preReportCallbacks []func()
	datapointCallbacks []DataPointCallback
	mu                 sync.Mutex
	oneShots           []DataPoint
	metricPrefix       string
}

// NewReporter returns a new Reporter object. Any dimensions supplied will be
// appended to all DataPoints sent to SignalFX. config is copied, so future
// changes to the external config object are not reflected within the reporter.
func NewReporter(config *Config,
	defaultDimensions map[string]string) *Reporter {
	return &Reporter{
		client:            NewClient(config),
		defaultDimensions: defaultDimensions,
		buckets:           map[*Bucket]interface{}{},
		metrics:           map[Metric]struct{}{},
	}
}

// SetPrefix sets a particular prefix for all metrics reported by this
// reporter.
func (r *Reporter) SetPrefix(prefix string) {
	r.lock()
	defer r.unlock()

	r.metricPrefix = prefix
}

func (r *Reporter) lock() {
	r.mu.Lock()
}

func (r *Reporter) unlock() {
	r.mu.Unlock()
}

// SetDimension sets a default dimension which will be reported for
// all data points.
func (r *Reporter) SetDimension(key, value string) {
	r.lock()
	defer r.unlock()

	r.defaultDimensions[key] = value
}

// DeleteDimension deletes a default dimension.
func (r *Reporter) DeleteDimension(key string) {
	r.lock()
	defer r.unlock()

	delete(r.defaultDimensions, key)
}

// Track adds a Metric to a Reporter's set of tracked Metrics.  Its
// value will be reported once each time Report is called.
func (r *Reporter) Track(m ...Metric) {
	r.lock()
	defer r.unlock()

	for _, m := range m {
		r.metrics[m] = struct{}{}
	}
	return
}

// Untrack removes a Metric from a Reporter's set of tracked Metrics.
func (r *Reporter) Untrack(m ...Metric) {
	r.lock()
	defer r.unlock()

	for _, m := range m {
		delete(r.metrics, m)
	}
}

// NewBucket creates a new Bucket object that is tracked by the Reporter.
// Buckets are goroutine safe.
func (r *Reporter) NewBucket(metric string, dimensions map[string]string) *Bucket {
	ret := NewBucket(r.metricPrefix+metric, dimensions)

	r.lock()
	defer r.unlock()

	r.buckets[ret] = nil
	return ret
}

// RemoveBucket takes Bucket(s) out of the set being tracked by the Reporter
func (r *Reporter) RemoveBucket(bs ...*Bucket) {
	r.lock()
	defer r.unlock()

	for _, b := range bs {
		delete(r.buckets, b)
	}
}

// AddPreReportCallback adds a function that is called before
// Report().  This is useful for refetching things like
// runtime.Memstats() so they are only fetched once per report()
// call. If a DataPoint
func (r *Reporter) AddPreReportCallback(f func()) {
	r.lock()
	defer r.unlock()
	r.preReportCallbacks = append(r.preReportCallbacks, f)
}

// AddDataPointsCallback adds a callback that itself will generate
// datapoints to report/
func (r *Reporter) AddDataPointsCallback(f DataPointCallback) {
	r.lock()
	defer r.unlock()
	r.datapointCallbacks = append(r.datapointCallbacks, f)
}

// Report sends all tracked DataPoints to SignalFX. PreReportCallbacks
// will be run before building the dataset to send. DataPoint
// callbacks will be executed and added to the dataset, but do not
// become tracked by the Reporter.
func (r *Reporter) Report(ctx context.Context) ([]DataPoint, error) {
	if ctx == nil {
		ctx = context.Background()
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.lock()
	defer r.unlock()

	dimensions := make([]*sfxproto.Dimension, 0, len(r.defaultDimensions))
	for k, v := range r.defaultDimensions {
		// have to copy the values, since these are stored as
		// pointers…
		var dk, dv string
		dk = k
		dv = v
		dimensions = append(dimensions,
			&sfxproto.Dimension{Key: &dk, Value: &dv})
	}

	for _, f := range r.preReportCallbacks {
		f()
	}

	// NOTE: yes, this assumes that there are five datapoints per
	// bucket.  This is normally true, for the normal bucket
	// use-case, and if it's false it just means either an extra
	// allocation or two one way or another.  Not a huge deal.
	//
	// It also assumes that there's a single data point returned
	// by each datapointCallback.  This can be altered if
	// experience shows otherwise.
	retLen := len(r.datapointCallbacks) + len(r.buckets)*5 +
		len(r.metrics) + len(r.oneShots)
	ret := make([]DataPoint, 0, retLen)

	for _, f := range r.datapointCallbacks {
		for _, dp := range f() {
			ret = append(ret, dp)
		}
	}

	for b := range r.buckets {
		ret = append(ret, b.DataPoints()...)
	}

	// append all of the one-shots
	for _, dp := range r.oneShots {
		ret = append(ret, dp)
	}

	hookedMetrics := make([]struct {
		m HookedMetric
		v int64
	}, 0)

	// append all of the tracked metrics
	for metric := range r.metrics {
		dp := metric.DataPoint()
		if dp == nil {
			continue
		}
		ret = append(ret, *dp)
		if m, ok := metric.(HookedMetric); ok {
			hm := struct {
				m HookedMetric
				v int64
			}{
				m,
				dp.Value,
			}
			hookedMetrics = append(hookedMetrics, hm)
		}
	}

	if len(ret) == 0 {
		return nil, nil
	}

	pdps := sfxproto.NewDataPoints(len(ret))
	for _, dp := range ret {
		pdp := dp.protoDataPoint(r.metricPrefix, dimensions)
		pdps.Add(pdp)
	}

	if err := r.client.Submit(ctx, pdps); err != nil {
		return nil, err
	}

	// reset resettable metrics
	for _, hm := range hookedMetrics {
		hm.m.PostReportHook(hm.v)
	}

	// and clear the one-shots
	r.oneShots = nil

	return ret, nil
}

func (r *Reporter) Add(dp DataPoint) {
	r.lock()
	defer r.unlock()

	r.oneShots = append(r.oneShots, dp)
}

// Inc adds a one-shot data point for a counter with the indicated
// delta since the last report.
func (r *Reporter) Inc(metric string, dimensions map[string]string, delta uint64) {
	if delta > math.MaxInt64 {
		log.Fatalf("counter increment %d is too large for int64", delta)
	}
	r.Add(DataPoint{
		Metric:     metric,
		Dimensions: dimensions,
		Type:       CounterType,
		Value:      int64(delta),
	})
}

// Record adds a one-shot data point for a gauge with the indicated
// value at this point in time.
func (r *Reporter) Record(metric string, dimensions map[string]string, value int64) {
	r.Add(DataPoint{
		Metric:     metric,
		Dimensions: dimensions,
		Type:       GaugeType,
		Value:      value,
	})
}

// Sample adds a one-shot data point for a cumulative counter with the
// indicated value at this point in time.
func (r *Reporter) Sample(metric string, dimensions map[string]string, value int64) {
	r.Add(DataPoint{
		Metric:     metric,
		Dimensions: dimensions,
		Type:       CumulativeCounterType,
		Value:      value,
	})
}

// RunInBackground starts a goroutine which calls Reporter.Report on
// the specified interval.  It returns a function which may be used to
// cancel the backgrounding.
func (r *Reporter) RunInBackground(interval time.Duration) (cancel func()) {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				_, err := r.Report(context.Background())
				if err != nil && err != sfxproto.ErrMarshalNoData {
					log.Print(err)
				}
			case <-done:
				return
			}
		}
	}()
	return func() {
		done <- struct{}{}
	}
}
