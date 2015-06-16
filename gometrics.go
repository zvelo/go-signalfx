package signalfx

import (
	"runtime"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// GoMetrics gathers and reports generally useful go system stats for the reporter
type GoMetrics struct {
	datapoints *DataPoints
	reporter   *Reporter
}

// NewGoMetrics registers the reporter to report go system metrics
func NewGoMetrics(reporter *Reporter) *GoMetrics {
	dims := sfxproto.Dimensions{
		"instance": "global_stats",
		"stattype": "golang_sys",
	}

	start := time.Now()
	mstat := runtime.MemStats{}
	ret := &GoMetrics{
		reporter: reporter,
	}

	ret.datapoints = NewDataPoints(30).
		Add(reporter.NewGauge("Alloc", Value(&mstat.Alloc), dims)).
		Add(reporter.NewCumulative("TotalAlloc", Value(&mstat.TotalAlloc), dims)).
		Add(reporter.NewGauge("Sys", Value(&mstat.Sys), dims)).
		Add(reporter.NewCumulative("Lookups", Value(&mstat.Lookups), dims)).
		Add(reporter.NewCumulative("Mallocs", Value(&mstat.Mallocs), dims)).
		Add(reporter.NewCumulative("Frees", Value(&mstat.Frees), dims)).
		Add(reporter.NewGauge("HeapAlloc", Value(&mstat.HeapAlloc), dims)).
		Add(reporter.NewGauge("HeapSys", Value(&mstat.HeapSys), dims)).
		Add(reporter.NewGauge("HeapIdle", Value(&mstat.HeapIdle), dims)).
		Add(reporter.NewGauge("HeapInuse", Value(&mstat.HeapInuse), dims)).
		Add(reporter.NewGauge("HeapReleased", Value(&mstat.HeapReleased), dims)).
		Add(reporter.NewGauge("HeapObjects", Value(&mstat.HeapObjects), dims)).
		Add(reporter.NewGauge("StackInuse", Value(&mstat.StackInuse), dims)).
		Add(reporter.NewGauge("StackSys", Value(&mstat.StackSys), dims)).
		Add(reporter.NewGauge("MSpanInuse", Value(&mstat.MSpanInuse), dims)).
		Add(reporter.NewGauge("MSpanSys", Value(&mstat.MSpanSys), dims)).
		Add(reporter.NewGauge("MCacheInuse", Value(&mstat.MCacheInuse), dims)).
		Add(reporter.NewGauge("MCacheSys", Value(&mstat.MCacheSys), dims)).
		Add(reporter.NewGauge("BuckHashSys", Value(&mstat.BuckHashSys), dims)).
		Add(reporter.NewGauge("GCSys", Value(&mstat.GCSys), dims)).
		Add(reporter.NewGauge("OtherSys", Value(&mstat.OtherSys), dims)).
		Add(reporter.NewGauge("NextGC", Value(&mstat.NextGC), dims)).
		Add(reporter.NewGauge("LastGC", Value(&mstat.LastGC), dims)).
		Add(reporter.NewCumulative("PauseTotalNs", Value(&mstat.PauseTotalNs), dims)).
		Add(reporter.NewGauge("NumGC", Value(&mstat.NumGC), dims)).
		Add(reporter.NewGauge("GOMAXPROCS", GetterFunc(func() (interface{}, error) { return runtime.GOMAXPROCS(0), nil }), dims)).
		Add(reporter.NewGauge("process.uptime.ns", GetterFunc(func() (interface{}, error) { return time.Now().Sub(start).Nanoseconds(), nil }), dims)).
		Add(reporter.NewGauge("num_cpu", GetterFunc(func() (interface{}, error) { return runtime.NumCPU(), nil }), dims)).
		Add(reporter.NewCumulative("num_cgo_call", GetterFunc(func() (interface{}, error) { return runtime.NumCgoCall(), nil }), dims)).
		Add(reporter.NewGauge("num_goroutine", GetterFunc(func() (interface{}, error) { return runtime.NumGoroutine(), nil }), dims))

	reporter.AddPreReportCallback(func() {
		runtime.ReadMemStats(&mstat)
	})

	return ret
}

// Close the metric source and will stop reporting these system stats to the
// reporter. Implements the io.Closer interface.
func (g *GoMetrics) Close() error {
	g.reporter.RemoveDataPoints(g.datapoints)
	return nil
}
