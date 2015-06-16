package signalfx

// TODO(jrubin) go back to atomic operations where possible

import (
	"math"
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// A Bucket trakcs groups of values, reporting the min/max as gauges, and
// count/sum/sum of squares as a cumulative counter. All operations on Buckets
// are goroutine safe.
type Bucket struct {
	metric       string
	dimensions   sfxproto.Dimensions
	count        int64
	min          int64
	max          int64
	sum          int64
	sumOfSquares int64
	mu           sync.Mutex
}

func (b *Bucket) lock() {
	b.mu.Lock()
}

func (b *Bucket) unlock() {
	b.mu.Unlock()
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
func (b *Bucket) Dimensions() sfxproto.Dimensions {
	b.lock()
	defer b.unlock()

	return b.dimensions.Clone()
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

// SetDimensions adds or overwrites multiple dimensions
func (b *Bucket) SetDimensions(dims sfxproto.Dimensions) {
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
func (b *Bucket) Count() int64 {
	b.lock()
	defer b.unlock()

	return b.count
}

// Min returns the lowest item added to the Bucket
func (b *Bucket) Min() int64 {
	b.lock()
	defer b.unlock()

	return b.min
}

// Max returns the highest item added to the Bucket
func (b *Bucket) Max() int64 {
	b.lock()
	defer b.unlock()

	return b.max
}

// Sum returns the sum of all items added to the Bucket
func (b *Bucket) Sum() int64 {
	b.lock()
	defer b.unlock()

	return b.sum
}

// SumOfSquares returns the sum of the square of all items added to the Bucket
func (b *Bucket) SumOfSquares() int64 {
	b.lock()
	defer b.unlock()

	return b.sumOfSquares
}

// NewBucket creates a new Bucket
func NewBucket(metric string, dimensions sfxproto.Dimensions) *Bucket {
	return &Bucket{
		metric:     metric,
		dimensions: dimensions,
	}
}

// Add an item to the Bucket, later reporting the result in the next report
// cycle.
func (b *Bucket) Add(val int64) {
	b.lock()
	defer b.unlock()

	b.count++
	b.sum += val
	b.sumOfSquares += val * val

	if b.count == 1 {
		b.min = val
		b.max = val
		return
	}

	if b.min > val {
		b.min = val
	}

	if b.max < val {
		b.max = val
	}
}

// depends on outer lock
func (b *Bucket) dimFor(defaultDims sfxproto.Dimensions, rollup string) sfxproto.Dimensions {
	dims := defaultDims.Append(b.dimensions)
	dims["rollup"] = rollup
	return dims
}

// CountDataPoint returns a DataPoint representing the Bucket's Count
func (b *Bucket) CountDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	dp, _ := NewCounter(b.metric, b.count, b.dimFor(defaultDims, "count"))
	return dp
}

// SumDataPoint returns a DataPoint representing the Bucket's Sum
func (b *Bucket) SumDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	dp, _ := NewCounter(b.metric, b.sum, b.dimFor(defaultDims, "sum"))
	return dp
}

// SumOfSquaresDataPoint returns a DataPoint representing the Bucket's
// SumOfSquares
func (b *Bucket) SumOfSquaresDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	dp, _ := NewCounter(b.metric, b.sumOfSquares, b.dimFor(defaultDims, "sumsquare"))
	return dp
}

// MinDataPoint returns a DataPoint representing the Bucket's Min. Note that
// this resets the Min value. nil is returned if no items have been added to the
// bucket since it was created or last reset.
func (b *Bucket) MinDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	var min int64
	b.min, min = math.MaxInt64, b.min

	if b.count != 0 && min != math.MaxInt64 {
		dp, _ := NewGauge(b.metric+".min", min, b.dimFor(defaultDims, "min"))
		return dp
	}

	return nil
}

// MaxDataPoint returns a DataPoint representing the Bucket's Max. Note that
// this resets the Max value. nil is returned if no items have been added to the
// bucket since it was created or last reset.
func (b *Bucket) MaxDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	var max int64
	b.max, max = math.MinInt64, b.max

	if b.count != 0 && max != math.MinInt64 {
		dp, _ := NewGauge(b.metric+".max", max, b.dimFor(defaultDims, "max"))
		return dp
	}

	return nil
}

// DataPoints returns a DataPoints object with DataPoint values for Count, Sum,
// SumOfSquares, Min and Max (if set). Note that this resets both the Min and
// Max values.
func (b *Bucket) DataPoints(defaultDims sfxproto.Dimensions) *DataPoints {
	return NewDataPoints(5).
		Add(b.CountDataPoint(defaultDims)).
		Add(b.SumDataPoint(defaultDims)).
		Add(b.SumOfSquaresDataPoint(defaultDims)).
		Add(b.MinDataPoint(defaultDims)).
		Add(b.MaxDataPoint(defaultDims))
}
