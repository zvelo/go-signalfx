package signalfx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
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

		So(len(datapoints), ShouldBeGreaterThan, 0)

		testDataPoint := func(dp DataPoint, t MetricType) {
			So(dp.Type, ShouldEqual, t)
			So(dp.Timestamp.Before(time.Now()), ShouldBeTrue)
			So(len(dp.Dimensions), ShouldEqual, 2)

			for key, value := range dp.Dimensions {
				switch key {
				case "instance":
					So(value, ShouldEqual, "global_stats")
				case "stattype":
					So(value, ShouldEqual, "golang_sys")
				default:
					So(value, ShouldEqual, forceFail)
				}
			}

			So(dp.Value, ShouldBeGreaterThanOrEqualTo, 0)
		}

		for _, dp := range datapoints {
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
				testDataPoint(dp, GaugeType)
			case "TotalAlloc", "Lookups", "Mallocs", "Frees", "PauseTotalNs", "num_cgo_call":
				testDataPoint(dp, CumulativeCounterType)
			default:
				So(dp.Metric, ShouldEqual, forceFail)
			}
		}

		So(gometrics.Close(), ShouldBeNil)
	})
}
