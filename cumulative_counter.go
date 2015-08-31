package signalfx

import (
	"math"
	"sync/atomic"
	"time"
)

type CumulativeCounter struct {
	metric               string
	dimensions           map[string]string
	value, previousValue uint64
}

func NewCumulativeCounter(metric string, dimensions map[string]string, value uint64) *CumulativeCounter {
	return &CumulativeCounter{metric: metric, dimensions: dimensions, value: value}
}

func (cc *CumulativeCounter) Sample(delta uint64) {
	atomic.StoreUint64(&cc.value, delta)
}

func (cc *CumulativeCounter) DataPoint() *DataPoint {
	previous := atomic.LoadUint64(&cc.previousValue)
	value := atomic.LoadUint64(&cc.value)
	if value == previous {
		return nil
	}
	if value > math.MaxInt64 {
		return nil
	}
	return &DataPoint{
		Metric:     cc.metric,
		Timestamp:  time.Now(),
		Type:       CumulativeCounterType,
		Dimensions: cc.dimensions,
		Value:      int64(value),
	}
}

func (cc *CumulativeCounter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative cumulative counter should be impossible")
	}
	vv := uint64(v)
	prev := atomic.LoadUint64(&cc.previousValue)
	if vv <= prev {
		return
	}
	for !atomic.CompareAndSwapUint64(&cc.previousValue, prev, vv) {
		prev = atomic.LoadUint64(&cc.previousValue)
		if vv <= prev {
			return
		}
	}
}

type WrappedCumulativeCounter struct {
	metric               string
	dimensions           map[string]string
	wrappedValue         Getter
	value, previousValue uint64
}

func WrapCumulativeCounter(
	metric string,
	dimensions map[string]string,
	value Getter,
) *WrappedCumulativeCounter {
	return &WrappedCumulativeCounter{
		metric:       metric,
		dimensions:   dimensions,
		wrappedValue: value,
	}
}

func (cc *WrappedCumulativeCounter) DataPoint() *DataPoint {
	previous := atomic.LoadUint64(&cc.previousValue)
	gottenValue, err := cc.wrappedValue.Get()
	if err != nil {
		return nil
	}
	value, err := toInt64(gottenValue)
	if err != nil {
		return nil
	}
	if value < 0 {
		return nil
	}
	if uint64(value) == previous {
		return nil
	}
	return &DataPoint{
		Metric:     cc.metric,
		Timestamp:  time.Now(),
		Type:       CumulativeCounterType,
		Dimensions: cc.dimensions,
		Value:      int64(value),
	}
}

func (cc *WrappedCumulativeCounter) PostReportHook(v int64) {
	if v < 0 {
		panic("negative cumulative counter should be impossible")
	}
	vv := uint64(v)
	prev := atomic.LoadUint64(&cc.previousValue)
	if vv <= prev {
		return
	}
	for !atomic.CompareAndSwapUint64(&cc.previousValue, prev, vv) {
		prev = atomic.LoadUint64(&cc.previousValue)
		if vv <= prev {
			return
		}
	}
}
