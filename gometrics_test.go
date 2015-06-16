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

		testDataPoint := func(dp *DataPoint, t sfxproto.MetricType) {
			So(dp.Type(), ShouldEqual, t)
			So(dp.Time().Before(time.Now()), ShouldBeTrue)
			So(len(dp.Dimensions()), ShouldEqual, 2)

			for key, value := range dp.Dimensions() {
				switch key {
				case "instance":
					So(value, ShouldEqual, "global_stats")
				case "stattype":
					So(value, ShouldEqual, "golang_sys")
				default:
					So(value, ShouldEqual, forceFail)
				}
			}

			So(dp.pdp.Value, ShouldNotBeNil)
			So(dp.StrValue(), ShouldEqual, "")
			So(dp.DoubleValue(), ShouldEqual, 0)
			So(dp.IntValue(), ShouldBeGreaterThanOrEqualTo, 0)
		}

		list := datapoints.List()
		for _, dp := range list {
			switch dp.Metric() {
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
				testDataPoint(dp, sfxproto.MetricType_GAUGE)
			case "TotalAlloc", "Lookups", "Mallocs", "Frees", "PauseTotalNs", "num_cgo_call":
				testDataPoint(dp, sfxproto.MetricType_CUMULATIVE_COUNTER)
			default:
				So(dp.Metric(), ShouldEqual, forceFail)
			}
		}

		So(gometrics.Close(), ShouldBeNil)
	})
}
