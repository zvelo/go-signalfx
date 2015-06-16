package signalfx

import (
	"math"
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// A Bucket trakcs groups of values, reporting the min/max as gauges, and
// count/sum/sum of squares as a cumulative counter
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

func (b *Bucket) Metric() string {
	b.lock()
	defer b.unlock()

	return b.metric
}

func (b *Bucket) SetMetric(name string) {
	b.lock()
	defer b.unlock()

	b.metric = name
}

func (b *Bucket) Dimensions() sfxproto.Dimensions {
	b.lock()
	defer b.unlock()

	return b.dimensions.Clone()
}

func (b *Bucket) SetDimension(key, value string) {
	b.lock()
	defer b.unlock()

	b.dimensions[key] = value
}

func (b *Bucket) SetDimensions(dims sfxproto.Dimensions) {
	b.lock()
	defer b.unlock()

	for key, value := range dims {
		b.dimensions[key] = value
	}
}

func (b *Bucket) RemoveDimension(keys ...string) {
	b.lock()
	defer b.unlock()

	for _, key := range keys {
		delete(b.dimensions, key)
	}
}

func (b *Bucket) Count() int64 {
	b.lock()
	defer b.unlock()

	return b.count
}

func (b *Bucket) Min() int64 {
	b.lock()
	defer b.unlock()

	return b.min
}

func (b *Bucket) Max() int64 {
	b.lock()
	defer b.unlock()

	return b.max
}

func (b *Bucket) Sum() int64 {
	b.lock()
	defer b.unlock()

	return b.sum
}

func (b *Bucket) SumOfSquares() int64 {
	b.lock()
	defer b.unlock()

	return b.sumOfSquares
}

func NewBucket(metric string, dimensions sfxproto.Dimensions) *Bucket {
	return &Bucket{
		metric:     metric,
		dimensions: dimensions,
	}
}

// Add an item to the bucket, later reporting the result in the next report
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
	dims := defaultDims.Concat(b.dimensions)
	dims["rollup"] = rollup
	return dims
}

func (b *Bucket) CountDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	dp, _ := NewCounter(b.metric, b.count, b.dimFor(defaultDims, "count"))
	return dp
}

func (b *Bucket) SumDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	dp, _ := NewCounter(b.metric, b.sum, b.dimFor(defaultDims, "sum"))
	return dp
}

func (b *Bucket) SumOfSquaresDataPoint(defaultDims sfxproto.Dimensions) *DataPoint {
	b.lock()
	defer b.unlock()

	dp, _ := NewCounter(b.metric, b.sumOfSquares, b.dimFor(defaultDims, "sumsquare"))
	return dp
}

// resets min
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

// resets max
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

// resets min and max
func (b *Bucket) DataPoints(defaultDims sfxproto.Dimensions) *DataPoints {
	return NewDataPoints(5).
		Add(b.CountDataPoint(defaultDims)).
		Add(b.SumDataPoint(defaultDims)).
		Add(b.SumOfSquaresDataPoint(defaultDims)).
		Add(b.MinDataPoint(defaultDims)).
		Add(b.MaxDataPoint(defaultDims))
}
