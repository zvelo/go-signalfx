package sfxgometrics

import (
	"runtime"
	"time"

	"github.com/zvelo/go-signalfx/sfxmetric"
	"github.com/zvelo/go-signalfx/sfxproto"
	"github.com/zvelo/go-signalfx/sfxreporter"
)

// GoMetrics gathers and reports generally useful go system stats for the reporter
type GoMetrics struct {
	metrics  *sfxmetric.Metrics
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

	ret.metrics = sfxmetric.NewMetrics(30).
		Add(reporter.Gauge("Alloc", sfxmetric.ValueGetter(&mstat.Alloc), dims)).
		Add(reporter.Cumulative("TotalAlloc", sfxmetric.ValueGetter(&mstat.TotalAlloc), dims)).
		Add(reporter.Gauge("Sys", sfxmetric.ValueGetter(&mstat.Sys), dims)).
		Add(reporter.Cumulative("Lookups", sfxmetric.ValueGetter(&mstat.Lookups), dims)).
		Add(reporter.Cumulative("Mallocs", sfxmetric.ValueGetter(&mstat.Mallocs), dims)).
		Add(reporter.Cumulative("Frees", sfxmetric.ValueGetter(&mstat.Frees), dims)).
		Add(reporter.Gauge("HeapAlloc", sfxmetric.ValueGetter(&mstat.HeapAlloc), dims)).
		Add(reporter.Gauge("HeapSys", sfxmetric.ValueGetter(&mstat.HeapSys), dims)).
		Add(reporter.Gauge("HeapIdle", sfxmetric.ValueGetter(&mstat.HeapIdle), dims)).
		Add(reporter.Gauge("HeapInuse", sfxmetric.ValueGetter(&mstat.HeapInuse), dims)).
		Add(reporter.Gauge("HeapReleased", sfxmetric.ValueGetter(&mstat.HeapReleased), dims)).
		Add(reporter.Gauge("HeapObjects", sfxmetric.ValueGetter(&mstat.HeapObjects), dims)).
		Add(reporter.Gauge("StackInuse", sfxmetric.ValueGetter(&mstat.StackInuse), dims)).
		Add(reporter.Gauge("StackSys", sfxmetric.ValueGetter(&mstat.StackSys), dims)).
		Add(reporter.Gauge("MSpanInuse", sfxmetric.ValueGetter(&mstat.MSpanInuse), dims)).
		Add(reporter.Gauge("MSpanSys", sfxmetric.ValueGetter(&mstat.MSpanSys), dims)).
		Add(reporter.Gauge("MCacheInuse", sfxmetric.ValueGetter(&mstat.MCacheInuse), dims)).
		Add(reporter.Gauge("MCacheSys", sfxmetric.ValueGetter(&mstat.MCacheSys), dims)).
		Add(reporter.Gauge("BuckHashSys", sfxmetric.ValueGetter(&mstat.BuckHashSys), dims)).
		Add(reporter.Gauge("GCSys", sfxmetric.ValueGetter(&mstat.GCSys), dims)).
		Add(reporter.Gauge("OtherSys", sfxmetric.ValueGetter(&mstat.OtherSys), dims)).
		Add(reporter.Gauge("NextGC", sfxmetric.ValueGetter(&mstat.NextGC), dims)).
		Add(reporter.Gauge("LastGC", sfxmetric.ValueGetter(&mstat.LastGC), dims)).
		Add(reporter.Cumulative("PauseTotalNs", sfxmetric.ValueGetter(&mstat.PauseTotalNs), dims)).
		Add(reporter.Gauge("NumGC", sfxmetric.ValueGetter(&mstat.NumGC), dims)).
		Add(reporter.Gauge("GOMAXPROCS", sfxmetric.GetterFunc(func() (interface{}, error) { return runtime.GOMAXPROCS(0), nil }), dims)).
		Add(reporter.Gauge("process.uptime.ns", sfxmetric.GetterFunc(func() (interface{}, error) { return time.Now().Sub(start).Nanoseconds(), nil }), dims)).
		Add(reporter.Gauge("num_cpu", sfxmetric.GetterFunc(func() (interface{}, error) { return runtime.NumCPU(), nil }), dims)).
		Add(reporter.Cumulative("num_cgo_call", sfxmetric.GetterFunc(func() (interface{}, error) { return runtime.NumCgoCall(), nil }), dims)).
		Add(reporter.Gauge("num_goroutine", sfxmetric.GetterFunc(func() (interface{}, error) { return runtime.NumGoroutine(), nil }), dims))

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
