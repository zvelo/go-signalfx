package signalfx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

func TestGoMetrics(t *testing.T) {
	const forceFail = false

	Convey("Testing GoMetrics", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`"OK"`))
		}))
		defer ts.Close()

		config := NewConfig()
		So(config, ShouldNotBeNil)

		config.URL = ts.URL

		reporter := NewReporter(config, nil)
		So(reporter, ShouldNotBeNil)

		gometrics := NewGoMetrics(reporter)

		datapoints, err := reporter.Report(context.Background())
		So(err, ShouldBeNil)
		So(datapoints, ShouldNotBeNil)

		So(datapoints.Len(), ShouldEqual, 30)

		testDataPoint := func(m *DataPoint, t sfxproto.MetricType) {
			So(m.Type(), ShouldEqual, t)
			So(m.Time().Before(time.Now()), ShouldBeTrue)
			So(len(m.Dimensions()), ShouldEqual, 2)

			for key, value := range m.Dimensions() {
				switch key {
				case "instance":
					So(value, ShouldEqual, "global_stats")
				case "stattype":
					So(value, ShouldEqual, "golang_sys")
				default:
					So(value, ShouldEqual, forceFail)
				}
			}

			So(m.dp.Value, ShouldNotBeNil)
			So(m.StrValue(), ShouldEqual, "")
			So(m.DoubleValue(), ShouldEqual, 0)
			So(m.IntValue(), ShouldBeGreaterThanOrEqualTo, 0)
		}

		list := datapoints.List()
		for _, m := range list {
			switch m.Metric() {
			case "Alloc",
				"Sys",
				"HeapAlloc",
				"HeapSys",
				"HeapIdle",
				"HeapInuse",
				"HeapReleased",
				"HeapObjects",
				"StackInuse",
				"StackSys",
				"MSpanInuse",
				"MSpanSys",
				"MCacheInuse",
				"MCacheSys",
				"BuckHashSys",
				"GCSys",
				"OtherSys",
				"NextGC",
				"LastGC",
				"NumGC",
				"GOMAXPROCS",
				"process.uptime.ns",
				"num_cpu",
				"num_goroutine":
				testDataPoint(m, sfxproto.MetricType_GAUGE)
			case "TotalAlloc", "Lookups", "Mallocs", "Frees", "PauseTotalNs", "num_cgo_call":
				testDataPoint(m, sfxproto.MetricType_CUMULATIVE_COUNTER)
			default:
				So(m.Metric(), ShouldEqual, forceFail)
			}
		}

		So(gometrics.Close(), ShouldBeNil)
	})
}
