package signalfx

import (
	"math"
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// A Bucket trakcs groups of values, reporting the min/max as gauges, and
// count/sum/sum of squares as a cumulative counter
type Bucket struct {
	Metric       string
	dimensions   sfxproto.Dimensions
	count        int64
	min          int64
	max          int64
	sum          int64
	sumOfSquares int64
	lock         sync.Mutex
}

func (b *Bucket) Count() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.count
}

func (b *Bucket) Min() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.min
}

func (b *Bucket) Max() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.max
}

func (b *Bucket) Sum() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.sum
}

func (b *Bucket) SumOfSquares() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.sumOfSquares
}

func NewBucket(metric string, dimensions sfxproto.Dimensions) *Bucket {
	return &Bucket{
		Metric:     metric,
		dimensions: dimensions,
	}
}

// Add an item to the bucket, later reporting the result in the next report
// cycle.
func (b *Bucket) Add(val int64) {
	b.lock.Lock()
	defer b.lock.Unlock()

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

func (b *Bucket) CountMetric(defaultDims sfxproto.Dimensions) *Metric {
	b.lock.Lock()
	defer b.lock.Unlock()

	dp, _ := NewCounter(b.Metric, b.count, b.dimFor(defaultDims, "count"))
	return dp
}

func (b *Bucket) SumMetric(defaultDims sfxproto.Dimensions) *Metric {
	b.lock.Lock()
	defer b.lock.Unlock()

	dp, _ := NewCounter(b.Metric, b.sum, b.dimFor(defaultDims, "sum"))
	return dp
}

func (b *Bucket) SumOfSquaresMetric(defaultDims sfxproto.Dimensions) *Metric {
	b.lock.Lock()
	defer b.lock.Unlock()

	dp, _ := NewCounter(b.Metric, b.sumOfSquares, b.dimFor(defaultDims, "sumsquare"))
	return dp
}

// resets min
func (b *Bucket) MinMetric(defaultDims sfxproto.Dimensions) *Metric {
	b.lock.Lock()
	defer b.lock.Unlock()

	var min int64
	b.min, min = math.MaxInt64, b.min

	if b.count != 0 && min != math.MaxInt64 {
		dp, _ := NewGauge(b.Metric+".min", min, b.dimFor(defaultDims, "min"))
		return dp
	}

	return nil
}

// resets max
func (b *Bucket) MaxMetric(defaultDims sfxproto.Dimensions) *Metric {
	b.lock.Lock()
	defer b.lock.Unlock()

	var max int64
	b.max, max = math.MinInt64, b.max

	if b.count != 0 && max != math.MinInt64 {
		dp, _ := NewGauge(b.Metric+".max", max, b.dimFor(defaultDims, "max"))
		return dp
	}

	return nil
}

// resets min and max
func (b *Bucket) Metrics(defaultDims sfxproto.Dimensions) *Metrics {
	return NewMetrics(5).
		Add(b.CountMetric(defaultDims)).
		Add(b.SumMetric(defaultDims)).
		Add(b.SumOfSquaresMetric(defaultDims)).
		Add(b.MinMetric(defaultDims)).
		Add(b.MaxMetric(defaultDims))
}
