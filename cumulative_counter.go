package signalfx

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

type CumulativeCounter struct {
	Metric               string
	Dimensions           map[string]string
	value, previousValue uint64
}

func (cc *CumulativeCounter) Inc(delta uint64) uint64 {
	return atomic.AddUint64(&cc.value, delta)
}

func (cc *CumulativeCounter) dataPoint() *dataPoint {
	previous := atomic.LoadUint64(&cc.previousValue)
	value := atomic.LoadUint64(&cc.value)
	if value == previous {
		return nil
	}
	if value > math.MaxInt64 {
		return nil
	}
	return &dataPoint{
		Metric:     cc.Metric,
		Timestamp:  time.Now(),
		Type:       CumulativeCounterType,
		Dimensions: cc.Dimensions,
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
	Metric               string
	Dimensions           map[string]string
	Value                Getter
	value, previousValue uint64
}

func (cc *WrappedCumulativeCounter) dataPoint() *dataPoint {
	previous := atomic.LoadUint64(&cc.previousValue)
	gottenValue, err := cc.Value.Get()
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	value, err := toInt64(gottenValue)
	if err != nil {
		fmt.Println("error2:", err)
		return nil
	}
	if value < 0 {
		return nil
	}
	if uint64(value) == previous {
		return nil
	}
	return &dataPoint{
		Metric:     cc.Metric,
		Timestamp:  time.Now(),
		Type:       CumulativeCounterType,
		Dimensions: cc.Dimensions,
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