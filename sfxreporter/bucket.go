package sfxreporter

import (
	"math"
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// A Bucket trakcs groups of values, reporting the min/max as gauges, and
// count/sum/sum of squares as a cumulative counter
type Bucket struct {
	metricName   string
	dimensions   sfxproto.Dimensions
	count        int64
	min          int64
	max          int64
	sum          int64
	sumOfSquares int64
	lock         sync.Mutex
}

func NewBucket(metricName string, dimensions sfxproto.Dimensions) *Bucket {
	return &Bucket{
		metricName: metricName,
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

func (b *Bucket) dimFor(defaultDims sfxproto.Dimensions, rollup string) sfxproto.Dimensions {
	dims := defaultDims.Concat(b.dimensions)
	dims["rollup"] = rollup
	return dims
}

func (b *Bucket) Count(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	dp, _ := sfxproto.NewCounter(b.metricName, b.count, b.dimFor(defaultDims, "count"))
	return dp
}

func (b *Bucket) Sum(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	dp, _ := sfxproto.NewCounter(b.metricName, b.sum, b.dimFor(defaultDims, "sum"))
	return dp
}

func (b *Bucket) SumOfSquares(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	dp, _ := sfxproto.NewCounter(b.metricName, b.sumOfSquares, b.dimFor(defaultDims, "sumsquare"))
	return dp
}

// resets min
func (b *Bucket) Min(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	var min int64
	b.min, min = math.MaxInt64, b.min

	if b.count != 0 && min != math.MaxInt64 {
		dp, _ := sfxproto.NewGauge(b.metricName+".min", min, b.dimFor(defaultDims, "min"))
		return dp
	}

	return nil
}

// resets max
func (b *Bucket) Max(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	var max int64
	b.max, max = math.MinInt64, b.max

	if b.count != 0 && max != math.MinInt64 {
		dp, _ := sfxproto.NewGauge(b.metricName+".max", max, b.dimFor(defaultDims, "max"))
		return dp
	}

	return nil
}

// resets min and max
func (b *Bucket) DataPoints(defaultDims sfxproto.Dimensions) *sfxproto.DataPoints {
	return sfxproto.NewDataPoints(5).
		Add(b.Count(defaultDims)).
		Add(b.Sum(defaultDims)).
		Add(b.SumOfSquares(defaultDims)).
		Add(b.Min(defaultDims)).
		Add(b.Max(defaultDims))
}
