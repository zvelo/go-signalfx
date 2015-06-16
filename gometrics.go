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
		Add(reporter.NewGauge("Alloc", ValueGetter(&mstat.Alloc), dims)).
		Add(reporter.NewCumulative("TotalAlloc", ValueGetter(&mstat.TotalAlloc), dims)).
		Add(reporter.NewGauge("Sys", ValueGetter(&mstat.Sys), dims)).
		Add(reporter.NewCumulative("Lookups", ValueGetter(&mstat.Lookups), dims)).
		Add(reporter.NewCumulative("Mallocs", ValueGetter(&mstat.Mallocs), dims)).
		Add(reporter.NewCumulative("Frees", ValueGetter(&mstat.Frees), dims)).
		Add(reporter.NewGauge("HeapAlloc", ValueGetter(&mstat.HeapAlloc), dims)).
		Add(reporter.NewGauge("HeapSys", ValueGetter(&mstat.HeapSys), dims)).
		Add(reporter.NewGauge("HeapIdle", ValueGetter(&mstat.HeapIdle), dims)).
		Add(reporter.NewGauge("HeapInuse", ValueGetter(&mstat.HeapInuse), dims)).
		Add(reporter.NewGauge("HeapReleased", ValueGetter(&mstat.HeapReleased), dims)).
		Add(reporter.NewGauge("HeapObjects", ValueGetter(&mstat.HeapObjects), dims)).
		Add(reporter.NewGauge("StackInuse", ValueGetter(&mstat.StackInuse), dims)).
		Add(reporter.NewGauge("StackSys", ValueGetter(&mstat.StackSys), dims)).
		Add(reporter.NewGauge("MSpanInuse", ValueGetter(&mstat.MSpanInuse), dims)).
		Add(reporter.NewGauge("MSpanSys", ValueGetter(&mstat.MSpanSys), dims)).
		Add(reporter.NewGauge("MCacheInuse", ValueGetter(&mstat.MCacheInuse), dims)).
		Add(reporter.NewGauge("MCacheSys", ValueGetter(&mstat.MCacheSys), dims)).
		Add(reporter.NewGauge("BuckHashSys", ValueGetter(&mstat.BuckHashSys), dims)).
		Add(reporter.NewGauge("GCSys", ValueGetter(&mstat.GCSys), dims)).
		Add(reporter.NewGauge("OtherSys", ValueGetter(&mstat.OtherSys), dims)).
		Add(reporter.NewGauge("NextGC", ValueGetter(&mstat.NextGC), dims)).
		Add(reporter.NewGauge("LastGC", ValueGetter(&mstat.LastGC), dims)).
		Add(reporter.NewCumulative("PauseTotalNs", ValueGetter(&mstat.PauseTotalNs), dims)).
		Add(reporter.NewGauge("NumGC", ValueGetter(&mstat.NumGC), dims)).
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
