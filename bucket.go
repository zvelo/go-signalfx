package signalfx

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/zvelo/go-signalfx/sfxproto"
)

var (
	// BucketMetricCount represents the count of datapoints seen
	BucketMetricCount = 1
	// BucketMetricMin represents the smallest datapoint seen
	BucketMetricMin = 2
	// BucketMetricMax represents the largest datapoint seen
	BucketMetricMax = 3
	// BucketMetricSum represents the sum of all datapoints seen
	BucketMetricSum = 4
	// BucketMetricSumOfSquares represents the sum of squares of all datapoints seen
	BucketMetricSumOfSquares = 5
)

// A Bucket trakcs groups of values, reporting metrics as gauges and
// resetting each time it reports. All operations on Buckets are goroutine safe.
type Bucket struct {
	metric       string
	dimensions   map[string]string
	count        uint64
	min          int64
	max          int64
	sum          int64
	sumOfSquares int64
	mu           sync.Mutex

	DisabledMetrics map[int]bool
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
		DisabledMetrics: b.DisabledMetrics,
	}
}

// Disable disables the given metrics for this bucket.
// They will be collected, but not reported.
func (b *Bucket) Disable(metrics ...int) {
	b.lock()
	defer b.unlock()
	for _, metric := range metrics {
		b.DisabledMetrics[metric] = true
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
		DisabledMetrics: make(map[int]bool, 0),
	}
}

// Add an item to the Bucket, later reporting the result in the next report
// cycle.
func (b *Bucket) Add(val int64) {
	// wrap all of these changes in the mutex lock since they need to occur as a
	// transaction, not a set of atomic operations
	b.lock()
	defer b.unlock()

	// still use atomic though, so that the atomic "getters" will never read
	// from an inconsistent state
	_ = atomic.AddUint64(&b.count, 1)
	atomic.AddInt64(&b.sum, val)
	atomic.AddInt64(&b.sumOfSquares, val*val)

	if b.Min() > val {
		atomic.StoreInt64(&b.min, val)
	}

	if b.Max() < val {
		atomic.StoreInt64(&b.max, val)
	}
}

func (b *Bucket) dimFor(defaultDims map[string]string, rollup string) map[string]string {
	b.lock()
	defer b.unlock()

	dims := sfxproto.Dimensions(defaultDims).Append(b.dimensions)
	dims["rollup"] = rollup

	return dims
}

// CountDataPoint returns a DataPoint representing the Bucket's Count. Because
// the passed in dimensions can not be locked by this method, it is important
// that the caller ensures its state does not change for the duration of the
// operation.
func (b *Bucket) CountDataPoint(defaultDims map[string]string) *DataPoint {
	cnt := atomic.SwapUint64(&b.count, 0)
	dp, _ := NewGauge(b.Metric(), cnt, b.dimFor(defaultDims, "count"))
	return dp
}

// SumDataPoint returns a DataPoint representing the Bucket's Sum. Because the
// passed in dimensions can not be locked by this method, it is important that
// the caller ensures its state does not change for the duration of the
// operation.
func (b *Bucket) SumDataPoint(defaultDims map[string]string) *DataPoint {
	sum := atomic.SwapInt64(&b.sum, 0)
	dp, _ := NewGauge(b.Metric(), sum, b.dimFor(defaultDims, "sum"))
	return dp
}

// SumOfSquaresDataPoint returns a DataPoint representing the Bucket's
// SumOfSquares. Because the passed in dimensions can not be locked by this
// method, it is important that the caller ensures its state does not change for
// the duration of the operation.
func (b *Bucket) SumOfSquaresDataPoint(defaultDims map[string]string) *DataPoint {
	sps := atomic.SwapInt64(&b.sumOfSquares, 0)
	dp, _ := NewGauge(b.Metric(), sps, b.dimFor(defaultDims, "sumsquare"))
	return dp
}

// MinDataPoint returns a DataPoint representing the Bucket's Min. Note that
// this resets the Min value. nil is returned if no items have been added to the
// bucket since it was created or last reset. Because the passed in dimensions
// can not be locked by this method, it is important that the caller ensures its
// state does not change for the duration of the operation.
func (b *Bucket) MinDataPoint(defaultDims map[string]string) *DataPoint {
	min := atomic.SwapInt64(&b.min, math.MaxInt64)

	if b.Count() != 0 && min != math.MaxInt64 {
		dp, _ := NewGauge(b.Metric(), min, b.dimFor(defaultDims, "min"))
		return dp
	}

	return nil
}

// MaxDataPoint returns a DataPoint representing the Bucket's Max. Note that
// this resets the Max value. nil is returned if no items have been added to the
// bucket since it was created or last reset. Because the passed in dimensions
// can not be locked by this method, it is important that the caller ensures its
// state does not change for the duration of the operation.
func (b *Bucket) MaxDataPoint(defaultDims map[string]string) *DataPoint {
	max := atomic.SwapInt64(&b.max, math.MinInt64)

	if b.Count() != 0 && max != math.MinInt64 {
		dp, _ := NewGauge(b.Metric(), max, b.dimFor(defaultDims, "max"))
		return dp
	}

	return nil
}

// DataPoints returns a DataPoints object with DataPoint values for Count, Sum,
// SumOfSquares, Min and Max (if set). Note that this resets all values.
// Because the passed in dimensions can not be locked by this method, it is
// important that the caller ensures its state does not change for the duration
// of the operation.
func (b *Bucket) DataPoints(defaultDims map[string]string) *DataPoints {
	dp := NewDataPoints(0)
	if !b.DisabledMetrics[BucketMetricMin] {
		dp = dp.Add(b.MinDataPoint(defaultDims))
	}
	if !b.DisabledMetrics[BucketMetricMax] {
		dp = dp.Add(b.MaxDataPoint(defaultDims))
	}
	if !b.DisabledMetrics[BucketMetricCount] {
		dp = dp.Add(b.CountDataPoint(defaultDims))
	}
	if !b.DisabledMetrics[BucketMetricSum] {
		dp = dp.Add(b.SumDataPoint(defaultDims))
	}
	if !b.DisabledMetrics[BucketMetricSumOfSquares] {
		dp = dp.Add(b.SumOfSquaresDataPoint(defaultDims))
	}

	return dp
}
