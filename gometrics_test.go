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
	const authToken = "abc123"
	const forceFail = false

	Convey("Testing GoMetrics", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`"OK"`))
		}))
		defer ts.Close()

		config := NewConfig(authToken)
		So(config, ShouldNotBeNil)

		config.URL = ts.URL

		reporter := NewReporter(config, nil)
		So(reporter, ShouldNotBeNil)

		gometrics := NewGoMetrics(reporter)

		dps, err := reporter.Report(context.Background())
		So(err, ShouldBeNil)
		So(dps, ShouldNotBeNil)

		So(dps.Len(), ShouldEqual, 30)

		testMetric := func(dp *sfxproto.DataPoint, t sfxproto.MetricType) {
			So(dp.MetricType, ShouldEqual, t)
			So(dp.Time().Before(time.Now()), ShouldBeTrue)
			So(len(dp.Dimensions), ShouldEqual, 2)

			for _, dp := range dp.Dimensions {
				switch dp.Key {
				case "instance":
					So(dp.Value, ShouldEqual, "global_stats")
				case "stattype":
					So(dp.Value, ShouldEqual, "golang_sys")
				default:
					So(dp.Value, ShouldEqual, forceFail)
				}
			}

			So(dp.Value, ShouldNotBeNil)
			So(dp.Value.StrValue, ShouldEqual, "")
			So(dp.Value.DoubleValue, ShouldEqual, 0)
			So(dp.Value.IntValue, ShouldBeGreaterThanOrEqualTo, 0)
		}

		list := dps.List()
		for _, dp := range list {
			switch dp.Metric {
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
				testMetric(dp, sfxproto.MetricType_GAUGE)
			case "TotalAlloc", "Lookups", "Mallocs", "Frees", "PauseTotalNs", "num_cgo_call":
				testMetric(dp, sfxproto.MetricType_CUMULATIVE_COUNTER)
			default:
				So(dp.Metric, ShouldEqual, forceFail)
			}
		}

		So(gometrics.Close(), ShouldBeNil)
	})
}
