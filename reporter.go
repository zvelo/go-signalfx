package signalfx

import (
	"sync"
	"time"

	"github.com/coreos/fleet/log"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

type Metric interface {
	dataPoint(string, []*sfxproto.Dimension) *sfxproto.DataPoint
}

type HookedMetric interface {
	Metric
	PostReportHook(reportedValue int64)
}

// DataPointCallback is a functional callback that can be passed to
// DataPointCallback as a way to have the caller calculate and return
// their own datapoints
type DataPointCallback func(defaultDims map[string]string) *DataPoints

// Reporter is an object that tracks DataPoints and manages a Client. It is the
// recommended way to send data to SignalFX.
type Reporter struct {
	client             *Client
	defaultDimensions  map[string]string
	datapoints         *DataPoints
	metrics            map[Metric]struct{}
	buckets            map[*Bucket]interface{}
	preReportCallbacks []func()
	datapointCallbacks []DataPointCallback
	mu                 sync.Mutex
	oneShots           []*sfxproto.DataPoint
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
		datapoints:        NewDataPoints(0),
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

// NewCumulative returns a new DataPoint object with type CUMULATIVE_COUNTER.
// val can be any type of int, float, string, nil, pointer to those types, or a
// Getter that returns any of those types. Literal pointers are copied by value.
// Getters that return pointer types should not have their value changed, unless
// atomically, when in a Reporter, except within a PreReportCallback, for
// goroutine safety.
func (r *Reporter) NewCumulative(metric string, val interface{}, dims map[string]string) *DataPoint {
	dp, _ := NewCumulative(r.metricPrefix+metric, val, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	r.datapoints.Add(dp)
	return dp
}

// NewGauge returns a new DataPoint object with type GAUGE. val can be any type
// of int, float, string, nil, pointer to those types, or a Getter that returns
// any of those types. Literal pointers are copied by value. Getters that return
// pointer types should not have their value changed, unless atomically, when in
// a Reporter, except within a PreReportCallback, for goroutine safety.
func (r *Reporter) NewGauge(metric string, val interface{}, dims map[string]string) *DataPoint {
	dp, _ := NewGauge(r.metricPrefix+metric, val, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	r.datapoints.Add(dp)
	return dp
}

// NewCounter returns a new DataPoint object with type COUNTER. val can be any
// type of int, float, string, nil, pointer to those types, or a Getter that
// returns any of those types. Literal pointers are copied by value. Getters
// that return pointer types should not have their value changed, unless
// atomically, when in a Reporter, except within a PreReportCallback, for
// goroutine safety.
func (r *Reporter) NewCounter(metric string, val interface{}, dims map[string]string) *DataPoint {
	dp, _ := NewCounter(r.metricPrefix+metric, val, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	r.datapoints.Add(dp)
	return dp
}

// NewInt32 creates a new DataPoint object with type COUNTER whose value is
// bound to an Int32. All methods on Int32 are goroutine safe and it may also be
// modified in atomic operations as an int32.
func (r *Reporter) NewInt32(metric string, dims map[string]string) (*Int32, *DataPoint) {
	ret := Int32(0)
	dp, _ := NewCounter(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewCumulativeInt32 creates a new DataPoint object with type
// CUMULATIVE_COUNTER whose value is bound to an Int32. All methods on Int32 are
// goroutine safe and it may also be modified in atomic operations as an int32.
func (r *Reporter) NewCumulativeInt32(metric string, dims map[string]string) (*Int32, *DataPoint) {
	ret := Int32(0)
	dp, _ := NewCumulative(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewInt64 creates a new DataPoint object with type COUNTER whose value is
// bound to an Int64. All methods on Int64 are goroutine safe and it may also be
// modified in atomic operations as an int64.
func (r *Reporter) NewInt64(metric string, dims map[string]string) (*Int64, *DataPoint) {
	ret := Int64(0)
	dp, _ := NewCounter(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewCumulativeInt64 creates a new DataPoint object with type
// CUMULATIVE_COUNTER whose value is bound to an Int64. All methods on Int64 are
// goroutine safe and it may also be modified in atomic operations as an int64.
func (r *Reporter) NewCumulativeInt64(metric string, dims map[string]string) (*Int64, *DataPoint) {
	ret := Int64(0)
	dp, _ := NewCumulative(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewUint32 creates a new DataPoint object with type COUNTER whose value is
// bound to an Uint32. All methods on Uint32 are goroutine safe and it may also
// be modified in atomic operations as an uint32.
func (r *Reporter) NewUint32(metric string, dims map[string]string) (*Uint32, *DataPoint) {
	ret := Uint32(0)
	dp, _ := NewCounter(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
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
func (r *Reporter) NewCumulativeUint32(metric string, dims map[string]string) (*Uint32, *DataPoint) {
	ret := Uint32(0)
	dp, _ := NewCumulative(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewUint64 creates a new DataPoint object with type COUNTER whose value is
// bound to an Uint64. All methods on Uint64 are goroutine safe and it may also be
// modified in atomic operations as an uint64.
func (r *Reporter) NewUint64(metric string, dims map[string]string) (*Uint64, *DataPoint) {
	ret := Uint64(0)
	dp, _ := NewCounter(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
	if dp == nil {
		return nil, nil
	}
	r.datapoints.Add(dp)
	return &ret, dp
}

// NewCumulativeUint64 creates a new DataPoint object with type
// CUMULATIVE_COUNTER whose value is bound to an Uint64. All methods on Uint64 are
// goroutine safe and it may also be modified in atomic operations as an uint64.
func (r *Reporter) NewCumulativeUint64(metric string, dims map[string]string) (*Uint64, *DataPoint) {
	ret := Uint64(0)
	dp, _ := NewCumulative(r.metricPrefix+metric, &ret, sfxproto.Dimensions(r.defaultDimensions).Append(dims))
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

	dimensions := make([]*sfxproto.Dimension, 0, len(r.defaultDimensions))
	for k, v := range r.defaultDimensions {
		// have to copy the values, since these are stored as
		// pointers…
		var dk, dv string
		dk = k
		dv = v
		dimensions = append(dimensions, &sfxproto.Dimension{Key: &dk, Value: &dv})
	}

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
				dp.SetTime(time.Now())
				return true
			}
		case sfxproto.MetricType_CUMULATIVE_COUNTER:
			if !dp.pdp.Equal(dp.previous) {
				cumulativeCounters = append(cumulativeCounters, dp)
				dp.SetTime(time.Now())
				return true
			}
		default:
			dp.SetTime(time.Now())
			return true
		}
		return false
	})

	pdps := ret.ProtoDataPoints()

	// append all of the one-shots
	for _, pdp := range r.oneShots {
		pdps.Add(pdp)
	}

	hookedMetrics := make([]struct {
		m  HookedMetric
		dp *sfxproto.DataPoint
	}, 0)
	// append all of the tracked metrics
	for metric := range r.metrics {
		dp := metric.dataPoint(r.metricPrefix, dimensions)
		if dp == nil {
			continue
		}
		pdps.Add(dp)
		if m, ok := metric.(HookedMetric); ok {
			rm := struct {
				m  HookedMetric
				dp *sfxproto.DataPoint
			}{
				m,
				dp,
			}
			hookedMetrics = append(hookedMetrics, rm)
		}
	}

	if err := r.client.Submit(ctx, pdps); err != nil {
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

	// reset resettable metrics
	for _, hm := range hookedMetrics {
		hm.m.PostReportHook(*hm.dp.Value.IntValue)
	}

	// and clear the one-shots
	r.oneShots = nil

	return ret, nil
}

// add wraps the one-shot submission process.
func (r *Reporter) add(metricType sfxproto.MetricType, metric string, dimensions map[string]string, value int64) error {
	r.lock()
	defer r.unlock()

	var protoDims []*sfxproto.Dimension
	for k, v := range r.defaultDimensions {
		// have to copy the values, since these are stored as
		// pointers…
		var dk, dv string
		dk = k
		dv = v
		protoDims = append(protoDims, &sfxproto.Dimension{Key: &dk, Value: &dv})
	}
	for k, v := range dimensions {
		// have to copy the values, since these are stored as
		// pointers…
		var dk, dv string
		dk = k
		dv = v
		protoDims = append(protoDims, &sfxproto.Dimension{Key: &dk, Value: &dv})
	}
	timestamp := time.Now().UnixNano() / 1000000
	// have to create a new addressable value, since metric is
	// stored as a pointer…
	prefixedMetric := r.metricPrefix + metric
	dp := &sfxproto.DataPoint{
		Metric:     &prefixedMetric,
		Timestamp:  &timestamp,
		MetricType: &metricType,
		Dimensions: protoDims,
		Value:      &sfxproto.Datum{IntValue: &value},
	}
	r.oneShots = append(r.oneShots, dp)
	return nil
}

// Inc adds a one-shot data point for a counter with the indicated
// delta since the last report.
func (r *Reporter) Inc(metric string, dimensions map[string]string, delta int64) error {
	return r.add(sfxproto.MetricType_COUNTER, metric, dimensions, delta)
}

// Record adds a one-shot data point for a gauge with the indicated
// value at this point in time.
func (r *Reporter) Record(metric string, dimensions map[string]string, value int64) error {
	return r.add(sfxproto.MetricType_GAUGE, metric, dimensions, value)
}

// Sample adds a one-shot data point for a cumulative counter with the
// indicated value at this point in time.
func (r *Reporter) Sample(metric string, dimensions map[string]string, value int64) error {
	return r.add(sfxproto.MetricType_CUMULATIVE_COUNTER, metric, dimensions, value)
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
					log.Error(err)
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
