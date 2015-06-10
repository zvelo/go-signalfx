package sfxgometrics

import (
	"runtime"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
	"github.com/zvelo/go-signalfx/sfxreporter"
)

// GoMetrics gathers and reports generally useful go system stats for the reporter
type GoMetrics struct {
	metrics  *sfxreporter.Metrics
	reporter *sfxreporter.Reporter
}

// New registers the reporter to report go system metrics
func New(reporter *sfxreporter.Reporter) *GoMetrics {
	dims := sfxproto.Dimensions{
		"instance": "global_stats",
		"stattype": "golang_sys",
	}

	start := time.Now()
	mstat := runtime.MemStats{}
	ret := &GoMetrics{
		reporter: reporter,
	}

	ret.metrics = sfxreporter.NewMetrics(30).
		Add(reporter.Gauge("Alloc", sfxreporter.ValueGetter(&mstat.Alloc), dims)).
		Add(reporter.Cumulative("TotalAlloc", sfxreporter.ValueGetter(&mstat.TotalAlloc), dims)).
		Add(reporter.Gauge("Sys", sfxreporter.ValueGetter(&mstat.Sys), dims)).
		Add(reporter.Cumulative("Lookups", sfxreporter.ValueGetter(&mstat.Lookups), dims)).
		Add(reporter.Cumulative("Mallocs", sfxreporter.ValueGetter(&mstat.Mallocs), dims)).
		Add(reporter.Cumulative("Frees", sfxreporter.ValueGetter(&mstat.Frees), dims)).
		Add(reporter.Gauge("HeapAlloc", sfxreporter.ValueGetter(&mstat.HeapAlloc), dims)).
		Add(reporter.Gauge("HeapSys", sfxreporter.ValueGetter(&mstat.HeapSys), dims)).
		Add(reporter.Gauge("HeapIdle", sfxreporter.ValueGetter(&mstat.HeapIdle), dims)).
		Add(reporter.Gauge("HeapInuse", sfxreporter.ValueGetter(&mstat.HeapInuse), dims)).
		Add(reporter.Gauge("HeapReleased", sfxreporter.ValueGetter(&mstat.HeapReleased), dims)).
		Add(reporter.Gauge("HeapObjects", sfxreporter.ValueGetter(&mstat.HeapObjects), dims)).
		Add(reporter.Gauge("StackInuse", sfxreporter.ValueGetter(&mstat.StackInuse), dims)).
		Add(reporter.Gauge("StackSys", sfxreporter.ValueGetter(&mstat.StackSys), dims)).
		Add(reporter.Gauge("MSpanInuse", sfxreporter.ValueGetter(&mstat.MSpanInuse), dims)).
		Add(reporter.Gauge("MSpanSys", sfxreporter.ValueGetter(&mstat.MSpanSys), dims)).
		Add(reporter.Gauge("MCacheInuse", sfxreporter.ValueGetter(&mstat.MCacheInuse), dims)).
		Add(reporter.Gauge("MCacheSys", sfxreporter.ValueGetter(&mstat.MCacheSys), dims)).
		Add(reporter.Gauge("BuckHashSys", sfxreporter.ValueGetter(&mstat.BuckHashSys), dims)).
		Add(reporter.Gauge("GCSys", sfxreporter.ValueGetter(&mstat.GCSys), dims)).
		Add(reporter.Gauge("OtherSys", sfxreporter.ValueGetter(&mstat.OtherSys), dims)).
		Add(reporter.Gauge("NextGC", sfxreporter.ValueGetter(&mstat.NextGC), dims)).
		Add(reporter.Gauge("LastGC", sfxreporter.ValueGetter(&mstat.LastGC), dims)).
		Add(reporter.Cumulative("PauseTotalNs", sfxreporter.ValueGetter(&mstat.PauseTotalNs), dims)).
		Add(reporter.Gauge("NumGC", sfxreporter.ValueGetter(&mstat.NumGC), dims)).
		Add(reporter.Gauge("GOMAXPROCS", sfxreporter.GetterFunc(func() interface{} { return runtime.GOMAXPROCS(0) }), dims)).
		Add(reporter.Gauge("process.uptime.ns", sfxreporter.GetterFunc(func() interface{} { return time.Now().Sub(start).Nanoseconds() }), dims)).
		Add(reporter.Gauge("num_cpu", sfxreporter.GetterFunc(func() interface{} { return runtime.NumCPU() }), dims)).
		Add(reporter.Cumulative("num_cgo_call", sfxreporter.GetterFunc(func() interface{} { return runtime.NumCgoCall() }), dims)).
		Add(reporter.Gauge("num_goroutine", sfxreporter.GetterFunc(func() interface{} { return runtime.NumGoroutine() }), dims))

	reporter.AddPreCollectCallback(func() {
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
