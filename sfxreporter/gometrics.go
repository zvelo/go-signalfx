package sfxreporter

import (
	"runtime"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// GolangMetrics gathers and reports generally useful golang system stats for the reporter
type GolangMetrics struct {
	metrics  *Metrics
	reporter *Reporter
}

// NewGolangMetrics registers the reporter to report golang system metrics
func NewGolangMetrics(reporter *Reporter) *GolangMetrics {
	dims := sfxproto.Dimensions{
		"instance": "global_stats",
		"stattype": "golang_sys",
	}

	start := time.Now()
	ret := &GolangMetrics{
		reporter: reporter,
	}

	mstat := runtime.MemStats{}

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
		Add(reporter.Gauge("GOMAXPROCS", GetterFunc(func() interface{} { return runtime.GOMAXPROCS(0) }), dims)).
		Add(reporter.Gauge("process.uptime.ns", GetterFunc(func() interface{} { return time.Now().Sub(start).Nanoseconds() }), dims)).
		Add(reporter.Gauge("num_cpu", GetterFunc(func() interface{} { return runtime.NumCPU() }), dims)).
		Add(reporter.Cumulative("num_cgo_call", GetterFunc(func() interface{} { return runtime.NumCgoCall() }), dims)).
		Add(reporter.Gauge("num_goroutine", GetterFunc(func() interface{} { return runtime.NumGoroutine() }), dims))

	reporter.AddPreCollectCallback(func() {
		runtime.ReadMemStats(&mstat)
	})

	return ret
}

// Close the metric source and will stop reporting these system stats to the
// reporter. Implements the io.Closer interface.
func (g *GolangMetrics) Close() error {
	g.reporter.RemoveMetrics(g.metrics)
	return nil
}
