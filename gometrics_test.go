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
			_, _ = w.Write([]byte(`"OK"`))
		}))
		defer ts.Close()

		config := NewConfig()
		So(config, ShouldNotBeNil)

		config.URL = ts.URL

		reporter := NewReporter(config, nil)
		So(reporter, ShouldNotBeNil)

		gometrics := NewGoMetrics(reporter, map[string]string{"system": "test"})
		datapoints, err := reporter.Report(context.Background())
		So(err, ShouldBeNil)
		So(datapoints, ShouldNotBeNil)
		// This can vary depending on if a GC cycle has run yet.
		So(len(datapoints), ShouldBeGreaterThanOrEqualTo, 28)

		testDataPoint := func(dp DataPoint, t MetricType) {
			So(dp.Type, ShouldEqual, t)
			So(dp.Timestamp.Before(time.Now()), ShouldBeTrue)

			So(len(dp.Dimensions), ShouldEqual, 1)
			So(dp.Dimensions["system"], ShouldEqual, "test")

			So(dp.Value, ShouldBeGreaterThanOrEqualTo, 0)
		}

		for _, dp := range datapoints {
			switch dp.Metric {
			case "go-metric-alloc",
				"go-metric-sys",
				"go-metric-heap-alloc",
				"go-metric-heap-sys",
				"go-metric-heap-idle",
				"go-metric-heap-in-use",
				"go-metric-heap-released",
				"go-metric-heap-objects",
				"go-metric-stack-in-use",
				"go-metric-stack-sys",
				"go-metric-mspan-in-use",
				"go-metric-mspan-sys",
				"go-metric-mcache-in-use",
				"go-metric-mcache-sys",
				"go-metric-buck-hash-sys",
				"go-metric-gc-sys",
				"go-metric-other-sys",
				"go-metric-next-gc",
				"go-metric-last-gc",
				"go-metric-num-gc",
				"go-metric-gomaxprocs",
				"go-metric-uptime-ns",
				"go-metric-num-cpu",
				"go-metric-num-goroutine":
				testDataPoint(dp, GaugeType)
			case "go-metric-total-alloc", "go-metric-lookups", "go-metric-mallocs", "go-metric-frees", "go-metric-pause-total-ns", "go-metric-num-cgo-call":
				testDataPoint(dp, CumulativeCounterType)
			default:
				So(dp.Metric, ShouldEqual, forceFail)
			}
		}

		So(gometrics.Close(), ShouldBeNil)
	})
}
