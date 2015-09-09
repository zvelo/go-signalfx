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

// NewGoMetrics registers the reporter to report go system metrics
func NewGoMetrics(reporter *Reporter) *GoMetrics {
	dims := map[string]string{
		"instance": "global_stats",
		"stattype": "golang_sys",
	}

	start := time.Now()
	mstat := runtime.MemStats{}
	ret := &GoMetrics{
		reporter: reporter,
	}

	ret.metrics = []Metric{
		WrapGauge("Alloc", dims, Value(&mstat.Alloc)),
		WrapCumulativeCounter(
			"TotalAlloc",
			dims,
			Value(&mstat.TotalAlloc),
		),
		WrapGauge("Sys", dims, Value(&mstat.Sys)),
		WrapCumulativeCounter("Lookups", dims, Value(&mstat.Lookups)),
		WrapCumulativeCounter("Mallocs", dims, Value(&mstat.Mallocs)),
		WrapCumulativeCounter("Frees", dims, Value(&mstat.Frees)),
		WrapGauge("HeapAlloc", dims, Value(&mstat.HeapAlloc)),
		WrapGauge("HeapSys", dims, Value(&mstat.HeapSys)),
		WrapGauge("HeapIdle", dims, Value(&mstat.HeapIdle)),
		WrapGauge("HeapInuse", dims, Value(&mstat.HeapInuse)),
		WrapGauge("HeapReleased", dims, Value(&mstat.HeapReleased)),
		WrapGauge("HeapObjects", dims, Value(&mstat.HeapObjects)),
		WrapGauge("StackInuse", dims, Value(&mstat.StackInuse)),
		WrapGauge("StackSys", dims, Value(&mstat.StackSys)),
		WrapGauge("MSpanInuse", dims, Value(&mstat.MSpanInuse)),
		WrapGauge("MSpanSys", dims, Value(&mstat.MSpanSys)),
		WrapGauge("MCacheInuse", dims, Value(&mstat.MCacheInuse)),
		WrapGauge("MCacheSys", dims, Value(&mstat.MCacheSys)),
		WrapGauge("BuckHashSys", dims, Value(&mstat.BuckHashSys)),
		WrapGauge("GCSys", dims, Value(&mstat.GCSys)),
		WrapGauge("OtherSys", dims, Value(&mstat.OtherSys)),
		WrapGauge("NextGC", dims, Value(&mstat.NextGC)),
		WrapGauge("LastGC", dims, Value(&mstat.LastGC)),
		WrapCumulativeCounter(
			"PauseTotalNs",
			dims,
			Value(&mstat.PauseTotalNs),
		),
		WrapGauge("NumGC", dims, Value(&mstat.NumGC)),

		WrapGauge(
			"GOMAXPROCS",
			dims,
			GetterFunc(func() (interface{}, error) {
				return runtime.GOMAXPROCS(0), nil
			}),
		),
		WrapGauge(
			"process.uptime.ns",
			dims,
			GetterFunc(func() (interface{}, error) {
				return time.Now().Sub(start).Nanoseconds(), nil
			}),
		),
		WrapGauge(
			"num_cpu",
			dims,
			GetterFunc(func() (interface{}, error) {
				return runtime.NumCPU(), nil
			}),
		),
		WrapCumulativeCounter(
			"num_cgo_call",
			dims,
			GetterFunc(func() (interface{}, error) {
				return runtime.NumCgoCall(), nil
			}),
		),
		WrapGauge(
			"num_goroutine",
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
