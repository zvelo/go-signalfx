package signalfx

import (
	"runtime"
	"time"
)

// GoMetrics gathers and reports generally useful go system stats for the reporter
type GoMetrics struct {
	metrics  []Metric
	reporter *Reporter
}

// NewGoMetrics registers the reporter to report go system metrics.
// You should provide enough dims to differentiate this set of metrics.
func NewGoMetrics(reporter *Reporter, dims map[string]string) *GoMetrics {
	start := time.Now()
	mstat := runtime.MemStats{}
	ret := &GoMetrics{
		reporter: reporter,
	}

	ret.metrics = []Metric{
		WrapGauge("go-metric-alloc", dims, Value(&mstat.Alloc)),
		WrapCumulativeCounter(
			"go-metric-total-alloc",
			dims,
			Value(&mstat.TotalAlloc),
		),
		WrapGauge("go-metric-sys", dims, Value(&mstat.Sys)),
		WrapCumulativeCounter("go-metric-lookups", dims, Value(&mstat.Lookups)),
		WrapCumulativeCounter("go-metric-mallocs", dims, Value(&mstat.Mallocs)),
		WrapCumulativeCounter("go-metric-frees", dims, Value(&mstat.Frees)),
		WrapGauge("go-metric-heap-alloc", dims, Value(&mstat.HeapAlloc)),
		WrapGauge("go-metric-heap-sys", dims, Value(&mstat.HeapSys)),
		WrapGauge("go-metric-heap-idle", dims, Value(&mstat.HeapIdle)),
		WrapGauge("go-metric-heap-in-use", dims, Value(&mstat.HeapInuse)),
		WrapGauge("go-metric-heap-released", dims, Value(&mstat.HeapReleased)),
		WrapGauge("go-metric-heap-objects", dims, Value(&mstat.HeapObjects)),
		WrapGauge("go-metric-stack-in-use", dims, Value(&mstat.StackInuse)),
		WrapGauge("go-metric-stack-sys", dims, Value(&mstat.StackSys)),
		WrapGauge("go-metric-mspan-in-use", dims, Value(&mstat.MSpanInuse)),
		WrapGauge("go-metric-mspan-sys", dims, Value(&mstat.MSpanSys)),
		WrapGauge("go-metric-mcache-in-use", dims, Value(&mstat.MCacheInuse)),
		WrapGauge("go-metric-mcache-sys", dims, Value(&mstat.MCacheSys)),
		WrapGauge("go-metric-buck-hash-sys", dims, Value(&mstat.BuckHashSys)),
		WrapGauge("go-metric-gc-sys", dims, Value(&mstat.GCSys)),
		WrapGauge("go-metric-other-sys", dims, Value(&mstat.OtherSys)),
		WrapGauge("go-metric-next-gc", dims, Value(&mstat.NextGC)),
		WrapGauge("go-metric-last-gc", dims, Value(&mstat.LastGC)),
		WrapCumulativeCounter(
			"go-metric-pause-total-ns",
			dims,
			Value(&mstat.PauseTotalNs),
		),
		WrapGauge("go-metric-num-gc", dims, Value(&mstat.NumGC)),

		WrapGauge(
			"go-metric-gomaxprocs",
			dims,
			GetterFunc(func() (interface{}, error) {
				return runtime.GOMAXPROCS(0), nil
			}),
		),
		WrapGauge(
			"go-metric-uptime-ns",
			dims,
			GetterFunc(func() (interface{}, error) {
				return time.Since(start).Nanoseconds(), nil
			}),
		),
		WrapGauge(
			"go-metric-num-cpu",
			dims,
			GetterFunc(func() (interface{}, error) {
				return runtime.NumCPU(), nil
			}),
		),
		WrapCumulativeCounter(
			"go-metric-num-cgo-call",
			dims,
			GetterFunc(func() (interface{}, error) {
				return runtime.NumCgoCall(), nil
			}),
		),
		WrapGauge(
			"go-metric-num-goroutine",
			dims,
			GetterFunc(func() (interface{}, error) {
				return runtime.NumGoroutine(), nil
			}),
		),
	}
	reporter.Track(ret.metrics...)

	reporter.AddPreReportCallback(func() {
		runtime.ReadMemStats(&mstat)
	})

	return ret
}

// Close the metric source and will stop reporting these system stats to the
// reporter. Implements the io.Closer interface.
func (g *GoMetrics) Close() error {
	g.reporter.Untrack(g.metrics...)
	return nil
}
