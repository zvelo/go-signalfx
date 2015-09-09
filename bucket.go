package signalfx

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

const (
	// BucketMetricCount represents the count of datapoints seen
	BucketMetricCount = iota
	// BucketMetricMin represents the smallest datapoint seen
	BucketMetricMin = iota
	// BucketMetricMax represents the largest datapoint seen
	BucketMetricMax = iota
	// BucketMetricSum represents the sum of all datapoints seen
	BucketMetricSum = iota
	// BucketMetricSumOfSquares represents the sum of squares of all datapoints seen
	BucketMetricSumOfSquares = iota
)

// A Bucket trakcs groups of values, reporting metrics as gauges and
// resetting each time it reports. All operations on Buckets are goroutine safe.
type Bucket struct {
	metric             string
	dimensions         map[string]string
	count              uint64
	countMetric        Counter
	min                int64
	minMetric          Gauge
	max                int64
	maxMetric          Gauge
	sum                int64
	sumMetric          Gauge
	sumOfSquares       int64
	sumOfSquaresMetric Gauge
	mu                 sync.Mutex
	disabledMetrics    map[int]bool
}

func (b *Bucket) lock() {
	b.mu.Lock()
}

func (b *Bucket) unlock() {
	b.mu.Unlock()
}

// Clone makes a deep copy of a Bucket
func (b *Bucket) Clone() *Bucket {
	// use the mutex to keep all operations part of the same transaction
	b.lock()
	defer b.unlock()

	return &Bucket{
		metric:          b.metric,                                  // can't use Metric() since we already have a lock
		dimensions:      sfxproto.Dimensions(b.dimensions).Clone(), // can't use Dimensions() since we already have a lock
		count:           b.Count(),
		min:             b.Min(),
		max:             b.Max(),
		sum:             b.Sum(),
		sumOfSquares:    b.SumOfSquares(),
		disabledMetrics: b.disabledMetrics,
	}
}

// Disable disables the given metrics for this bucket.
// They will be collected, but not reported.
func (b *Bucket) Disable(metrics ...int) {
	b.lock()
	defer b.unlock()
	for _, metric := range metrics {
		b.disabledMetrics[metric] = true
	}
}

// Equal returns whether two buckets are exactly equal
func (b *Bucket) Equal(r *Bucket) bool {
	// lock the state of both buckets as this operation is effectively a
	// transaction on an exact state of a whole bucket

	b.lock()
	defer b.unlock()

	r.lock()
	defer r.unlock()

	if b.metric != r.metric {
		return false
	}

	if !sfxproto.Dimensions(b.dimensions).Equal(r.dimensions) {
		return false
	}

	if b.Count() != r.Count() {
		return false
	}

	if b.Min() != r.Min() {
		return false
	}

	if b.Max() != r.Max() {
		return false
	}

	if b.Sum() != r.Sum() {
		return false
	}

	if b.SumOfSquares() != r.SumOfSquares() {
		return false
	}

	return true
}

// Metric returns the metric name of the Bucket
func (b *Bucket) Metric() string {
	b.lock()
	defer b.unlock()

	return b.metric
}

// SetMetric sets the metric name of the Bucket
func (b *Bucket) SetMetric(name string) {
	b.lock()
	defer b.unlock()

	b.metric = name
}

// Dimensions returns a copy of the dimensions of the Bucket. Changes are not
// reflected inside the Bucket itself.
func (b *Bucket) Dimensions() map[string]string {
	b.lock()
	defer b.unlock()

	return sfxproto.Dimensions(b.dimensions).Clone()
}

// SetDimension adds or overwrites the dimension at key with value. If the key
// or value is empty, no changes are made
func (b *Bucket) SetDimension(key, value string) {
	if key == "" || value == "" {
		return
	}

	b.lock()
	defer b.unlock()

	b.dimensions[key] = value
}

// SetDimensions adds or overwrites multiple dimensions. Because the passed in
// dimensions can not be locked by this method, it is important that the caller
// ensures its state does not change for the duration of the operation.
func (b *Bucket) SetDimensions(dims map[string]string) {
	for key, value := range dims {
		b.SetDimension(key, value)
	}
}

// RemoveDimension removes one or more dimensions with the given keys
func (b *Bucket) RemoveDimension(keys ...string) {
	b.lock()
	defer b.unlock()

	for _, key := range keys {
		delete(b.dimensions, key)
	}
}

// Count returns the number of items added to the Bucket
func (b *Bucket) Count() uint64 {
	return atomic.LoadUint64(&b.count)
}

// Min returns the lowest item added to the Bucket
func (b *Bucket) Min() int64 {
	return atomic.LoadInt64(&b.min)
}

// Max returns the highest item added to the Bucket
func (b *Bucket) Max() int64 {
	return atomic.LoadInt64(&b.max)
}

// Sum returns the sum of all items added to the Bucket
func (b *Bucket) Sum() int64 {
	return atomic.LoadInt64(&b.sum)
}

// SumOfSquares returns the sum of the square of all items added to the Bucket
func (b *Bucket) SumOfSquares() int64 {
	return atomic.LoadInt64(&b.sumOfSquares)
}

// NewBucket creates a new Bucket. Because the passed in dimensions can not be
// locked by this method, it is important that the caller ensures its state does
// not change for the duration of the operation.
func NewBucket(metric string, dimensions map[string]string) *Bucket {
	return &Bucket{
		metric:          metric,
		dimensions:      sfxproto.Dimensions(dimensions).Clone(),
		min:             math.MaxInt64,
		max:             math.MinInt64,
		disabledMetrics: make(map[int]bool, 0),
	}
}

// Add an item to the Bucket, later reporting the result in the next report
// cycle.
func (b *Bucket) Add(val int64) {
	// still use atomic though, so that the atomic "getters" will never read
	// from an inconsistent state
	_ = atomic.AddUint64(&b.count, 1)
	atomic.AddInt64(&b.sum, val)
	atomic.AddInt64(&b.sumOfSquares, val*val)

	b.setIfMin(val)
	b.setIfMax(val)
}

func (b *Bucket) setIfMin(val int64) {
	for {
		if cur := b.Min(); cur > val || cur == math.MaxInt64 {
			if atomic.CompareAndSwapInt64(&b.min, cur, val) {
				break
			}
		} else {
			break
		}
	}
}

func (b *Bucket) setIfMax(val int64) {
	for {
		if cur := b.Max(); cur < val || cur == math.MinInt64 {
			if atomic.CompareAndSwapInt64(&b.max, cur, val) {
				break
			}
		} else {
			break
		}
	}
}

func (b *Bucket) dimFor(rollup string) map[string]string {
	b.lock()
	defer b.unlock()

	return sfxproto.Dimensions(map[string]string{"rollup": rollup}).Append(b.dimensions)
}

// DataPoints returns a DataPoints object with DataPoint values for
// Count, Sum, SumOfSquares, Min and Max (if set). Note that this
// resets all values.  If no values have been added to the bucket
// since the last report, it returns 0 for count, sum and
// sum-of-squares, omitting max and min.  If the count is higher than
// may be represented in an int64, then the count will be omitted.
func (b *Bucket) DataPoints() []DataPoint {
	dps := make([]DataPoint, 0, 5)
	cnt := atomic.SwapUint64(&b.count, 0)
	min := atomic.SwapInt64(&b.min, math.MaxInt64)
	max := atomic.SwapInt64(&b.max, math.MinInt64)
	sum := atomic.SwapInt64(&b.sum, 0)
	sos := atomic.SwapInt64(&b.sumOfSquares, 0)
	timestamp := time.Now()
	if cnt != 0 {
		if !b.disabledMetrics[BucketMetricMin] {
			dp := DataPoint{
				Metric:     b.metric,
				Dimensions: b.dimFor("min"),
				Type:       GaugeType,
				Value:      min,
				Timestamp:  timestamp,
			}
			dps = append(dps, dp)
		}
		if !b.disabledMetrics[BucketMetricMax] {
			dp := DataPoint{
				Metric:     b.metric,
				Dimensions: b.dimFor("max"),
				Type:       GaugeType,
				Value:      max,
				Timestamp:  timestamp,
			}
			dps = append(dps, dp)
		}
	}
	if !b.disabledMetrics[BucketMetricCount] && cnt <= math.MaxInt64 {
		dp := DataPoint{
			Metric:     b.metric,
			Dimensions: b.dimFor("count"),
			Type:       CounterType,
			Value:      int64(cnt),
			Timestamp:  timestamp,
		}
		dps = append(dps, dp)
	}
	if !b.disabledMetrics[BucketMetricSum] {
		dp := DataPoint{
			Metric:     b.metric,
			Dimensions: b.dimFor("sum"),
			Type:       GaugeType,
			Value:      sum,
			Timestamp:  timestamp,
		}
		dps = append(dps, dp)
	}
	if !b.disabledMetrics[BucketMetricSumOfSquares] {
		dp := DataPoint{
			Metric:     b.metric,
			Dimensions: b.dimFor("sumofsquares"),
			Type:       GaugeType,
			Value:      sos,
			Timestamp:  timestamp,
		}
		dps = append(dps, dp)
	}

	return dps
}
