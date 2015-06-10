package sfxreporter

import (
	"math"
	"sync"
	"time"

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
	time         time.Time
	lock         sync.Mutex
}

func NewBucket(metricName string, dimensions sfxproto.Dimensions) *Bucket {
	return &Bucket{
		metricName: metricName,
		dimensions: dimensions,
		time:       time.Now(),
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
	} else {
		if b.min > val {
			b.min = val
		}
		if b.max < val {
			b.max = val
		}
	}
}

func (b *Bucket) Merge(val *Bucket) {
	val.lock.Lock()
	defer val.lock.Unlock()

	if val.count == 0 {
		return
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	if b.metricName != val.metricName {
		return
	}

	b.count += val.count
	b.sum += val.sum
	b.sumOfSquares += val.sumOfSquares

	if val.min < b.min {
		b.min = val.min
	}

	if val.max > b.max {
		b.max = val.max
	}

	b.dimensions = append(b.dimensions, val.dimensions...)
}

func (b *Bucket) dimFor(defaultDims sfxproto.Dimensions, rollup string) sfxproto.Dimensions {
	dims := append(defaultDims, b.dimensions...)
	return append(dims, sfxproto.NewDimension("rollup", rollup))
}

func (b *Bucket) Count(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	return sfxproto.NewDataPoint(sfxproto.MetricType_COUNTER, b.metricName, b.count, b.time, b.dimFor(defaultDims, "count"))
}

func (b *Bucket) Sum(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	return sfxproto.NewDataPoint(sfxproto.MetricType_COUNTER, b.metricName, b.sum, b.time, b.dimFor(defaultDims, "sum"))
}

func (b *Bucket) SumOfSquares(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	return sfxproto.NewDataPoint(sfxproto.MetricType_COUNTER, b.metricName, b.sumOfSquares, b.time, b.dimFor(defaultDims, "sumsquare"))
}

// resets min
func (b *Bucket) Min(defaultDims sfxproto.Dimensions) *sfxproto.DataPoint {
	b.lock.Lock()
	defer b.lock.Unlock()

	var min int64
	b.min, min = math.MaxInt64, b.min

	if b.count != 0 && min != math.MaxInt64 {
		return sfxproto.NewDataPoint(sfxproto.MetricType_GAUGE, b.metricName+".min", min, b.time, b.dimFor(defaultDims, "min"))
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
		return sfxproto.NewDataPoint(sfxproto.MetricType_GAUGE, b.metricName+".max", max, b.time, b.dimFor(defaultDims, "max"))
	}

	return nil
}

// resets min and max
func (b *Bucket) DataPoints(defaultDims sfxproto.Dimensions) *sfxproto.DataPoints {
	return sfxproto.NewDataPoints(0, 5).
		Add(b.Count(defaultDims)).
		Add(b.Sum(defaultDims)).
		Add(b.SumOfSquares(defaultDims)).
		Add(b.Min(defaultDims)).
		Add(b.Max(defaultDims))
}
