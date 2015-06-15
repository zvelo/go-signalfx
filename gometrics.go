package signalfx

import (
	"runtime"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// GoMetrics gathers and reports generally useful go system stats for the reporter
type GoMetrics struct {
	metrics  *Metrics
	reporter *Reporter
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

	ret.metrics = NewMetrics(30).
		Add(reporter.Gauge("Alloc", ValueGetter(&mstat.Alloc), dims)).
		Add(reporter.Cumulative("TotalAlloc", ValueGetter(&mstat.TotalAlloc), dims)).
		Add(reporter.Gauge("Sys", ValueGetter(&mstat.Sys), dims)).
		Add(reporter.Cumulative("Lookups", ValueGetter(&mstat.Lookups), dims)).
		Add(reporter.Cumulative("Mallocs", ValueGetter(&mstat.Mallocs), dims)).
		Add(reporter.Cumulative("Frees", ValueGetter(&mstat.Frees), dims)).
		Add(reporter.Gauge("HeapAlloc", ValueGetter(&mstat.HeapAlloc), dims)).
		Add(reporter.Gauge("HeapSys", ValueGetter(&mstat.HeapSys), dims)).
		Add(reporter.Gauge("HeapIdle", ValueGetter(&mstat.HeapIdle), dims)).
		Add(reporter.Gauge("HeapInuse", ValueGetter(&mstat.HeapInuse), dims)).
		Add(reporter.Gauge("HeapReleased", ValueGetter(&mstat.HeapReleased), dims)).
		Add(reporter.Gauge("HeapObjects", ValueGetter(&mstat.HeapObjects), dims)).
		Add(reporter.Gauge("StackInuse", ValueGetter(&mstat.StackInuse), dims)).
		Add(reporter.Gauge("StackSys", ValueGetter(&mstat.StackSys), dims)).
		Add(reporter.Gauge("MSpanInuse", ValueGetter(&mstat.MSpanInuse), dims)).
		Add(reporter.Gauge("MSpanSys", ValueGetter(&mstat.MSpanSys), dims)).
		Add(reporter.Gauge("MCacheInuse", ValueGetter(&mstat.MCacheInuse), dims)).
		Add(reporter.Gauge("MCacheSys", ValueGetter(&mstat.MCacheSys), dims)).
		Add(reporter.Gauge("BuckHashSys", ValueGetter(&mstat.BuckHashSys), dims)).
		Add(reporter.Gauge("GCSys", ValueGetter(&mstat.GCSys), dims)).
		Add(reporter.Gauge("OtherSys", ValueGetter(&mstat.OtherSys), dims)).
		Add(reporter.Gauge("NextGC", ValueGetter(&mstat.NextGC), dims)).
		Add(reporter.Gauge("LastGC", ValueGetter(&mstat.LastGC), dims)).
		Add(reporter.Cumulative("PauseTotalNs", ValueGetter(&mstat.PauseTotalNs), dims)).
		Add(reporter.Gauge("NumGC", ValueGetter(&mstat.NumGC), dims)).
		Add(reporter.Gauge("GOMAXPROCS", GetterFunc(func() (interface{}, error) { return runtime.GOMAXPROCS(0), nil }), dims)).
		Add(reporter.Gauge("process.uptime.ns", GetterFunc(func() (interface{}, error) { return time.Now().Sub(start).Nanoseconds(), nil }), dims)).
		Add(reporter.Gauge("num_cpu", GetterFunc(func() (interface{}, error) { return runtime.NumCPU(), nil }), dims)).
		Add(reporter.Cumulative("num_cgo_call", GetterFunc(func() (interface{}, error) { return runtime.NumCgoCall(), nil }), dims)).
		Add(reporter.Gauge("num_goroutine", GetterFunc(func() (interface{}, error) { return runtime.NumGoroutine(), nil }), dims))

	reporter.AddPreReportCallback(func() {
		runtime.ReadMemStats(&mstat)
	})

	return ret
}

// Close the metric source and will stop reporting these system stats to the
// reporter. Implements the io.Closer interface.
func (g *GoMetrics) Close() error {
	g.reporter.RemoveMetrics(g.metrics)
	return nil
}
