package signalfx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

func TestReporter(t *testing.T) {
	Convey("Testing Reporter", t, func(c C) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.So(r.Header.Get(TokenHeader), ShouldEqual, "abcdefg")
			w.Write([]byte(`"OK"`))
		}))
		defer ts.Close()

		config := NewConfig()
		So(config, ShouldNotBeNil)

		config.URL = ts.URL
		config.AuthToken = "abcdefg"

		reporter := NewReporter(config, nil)
		So(reporter, ShouldNotBeNil)

		So(reporter.datapoints.Len(), ShouldEqual, 0)
		So(len(reporter.buckets), ShouldEqual, 0)

		Convey("working with datapoints", func() {
			// getting datapoints

			bucket := reporter.NewBucket("bucket", nil)
			So(bucket, ShouldNotBeNil)
			So(reporter.datapoints.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			cumulative := reporter.NewCumulative("cumulative", Value(0), nil)
			So(cumulative, ShouldNotBeNil)
			So(reporter.datapoints.Len(), ShouldEqual, 1)
			So(len(reporter.buckets), ShouldEqual, 1)

			gauge := reporter.NewGauge("gauge", Value(0), nil)
			So(gauge, ShouldNotBeNil)
			So(reporter.datapoints.Len(), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			counter := reporter.NewCounter("counter", Value(0), nil)
			So(counter, ShouldNotBeNil)
			So(reporter.datapoints.Len(), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 1)

			// removing datapoints

			reporter.RemoveDataPoint(cumulative)
			So(reporter.datapoints.Len(), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.RemoveDataPoint(cumulative)
			So(reporter.datapoints.Len(), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.RemoveDataPoint(gauge, counter)
			So(reporter.datapoints.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.AddDataPoint(gauge)
			So(reporter.datapoints.Len(), ShouldEqual, 1)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.AddDataPoint(cumulative, counter, gauge)
			So(reporter.datapoints.Len(), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 1)

			datapoints := NewDataPoints(3)
			datapoints.Add(cumulative, gauge, counter)
			So(datapoints.Len(), ShouldEqual, 3)

			reporter.RemoveDataPoints(datapoints)
			So(reporter.datapoints.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.RemoveBucket(bucket)
			So(reporter.datapoints.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 0)

			reporter.AddDataPoints(datapoints)
			So(reporter.datapoints.Len(), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 0)
		})

		Convey("callbacks should be fired", func() {
			So(reporter.datapoints.Len(), ShouldEqual, 0)

			cb := 0
			addDataPointF := func(dims sfxproto.Dimensions) *DataPoints {
				cb++
				count0, err := NewCounter("count0", Value(0), nil)
				if err != nil {
					return nil
				}
				So(count0, ShouldNotBeNil)

				count1, err := NewCounter("count1", Value(0), nil)
				if err != nil {
					return nil
				}

				So(count1, ShouldNotBeNil)
				return NewDataPoints(2).
					Add(count0).
					Add(count1)
			}

			addDataPointErrF := func(dims sfxproto.Dimensions) *DataPoints {
				cb++
				return nil
			}

			reporter.AddPreReportCallback(func() { cb++ })
			reporter.AddDataPointsCallback(addDataPointF)
			reporter.AddDataPointsCallback(addDataPointErrF)

			dps, err := reporter.Report(nil)
			So(err, ShouldBeNil)
			So(dps, ShouldNotBeNil)
			So(dps.Len(), ShouldEqual, 2)
			So(reporter.datapoints.Len(), ShouldEqual, 0)
			So(cb, ShouldEqual, 3)
		})

		Convey("reporting should work", func() {
			reporter.NewBucket("bucket", nil)
			dps, err := reporter.Report(context.Background())
			So(err, ShouldBeNil)
			So(dps.Len(), ShouldEqual, 3)
		})

		Convey("report should handle a previously canceled context", func() {
			// test report with already canceled context
			ctx, cancelF := context.WithCancel(context.Background())

			cancelF()
			<-ctx.Done()

			dps, err := reporter.Report(ctx)
			So(err, ShouldNotBeNil)
			So(dps, ShouldBeNil)
			So(err.Error(), ShouldEqual, "context canceled")
		})

		Convey("report should handle an 'in-flight' context cancellation", func() {
			reporter.NewBucket("bucket", nil)
			ctx, cancelF := context.WithCancel(context.Background())
			go cancelF()
			dps, err := reporter.Report(ctx)
			So(err, ShouldNotBeNil)
			So(dps, ShouldBeNil)
			So(err.Error(), ShouldEqual, "context canceled")
		})

		Convey("no metrics", func() {
			dps, err := reporter.Report(context.Background())
			So(err, ShouldNotBeNil)
			So(dps, ShouldBeNil)
			So(err.Error(), ShouldEqual, "no data to marshal")
		})

		Convey("report should fail with a bad url", func() {
			ccopy := config.Clone()
			ccopy.URL = "z" + ts.URL
			tmpR := NewReporter(ccopy, nil)
			tmpR.NewBucket("bucket", nil)
			dps, err := tmpR.Report(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Post z"+ts.URL+": unsupported protocol scheme \"zhttp\"")
			So(dps, ShouldBeNil)
		})

		Convey("incrementers should work", func() {
			inc, dp := reporter.NewInc("Incrementer", nil)
			So(dp, ShouldNotBeNil)
			So(inc, ShouldNotBeNil)
			So(reporter.datapoints.Len(), ShouldEqual, 1)
			So(len(reporter.buckets), ShouldEqual, 0)

			So(inc.Value(), ShouldEqual, 0)
			v, err := inc.Get()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 0)
			So(inc.Value(), ShouldEqual, dp.IntValue())

			inc.Set(5)
			So(inc.Value(), ShouldEqual, 5)
			v, err = inc.Get()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 5)
			So(inc.Value(), ShouldEqual, dp.IntValue())

			inc.Inc(1)
			So(inc.Value(), ShouldEqual, 6)
			v, err = inc.Get()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 6)
			So(inc.Value(), ShouldEqual, dp.IntValue())

			inc, dp = reporter.NewCumulativeInc("CumulativeIncrementer", nil)
			So(inc, ShouldNotBeNil)
			So(reporter.datapoints.Len(), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 0)

			So(inc.Value(), ShouldEqual, 0)
			v, err = inc.Get()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 0)
			So(inc.Value(), ShouldEqual, dp.IntValue())

			inc.Set(5)
			So(inc.Value(), ShouldEqual, 5)
			v, err = inc.Get()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 5)
			So(inc.Value(), ShouldEqual, dp.IntValue())

			inc.Inc(1)
			So(inc.Value(), ShouldEqual, 6)
			v, err = inc.Get()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 6)
			So(inc.Value(), ShouldEqual, dp.IntValue())
		})
	})
}

func ExampleReporter() {
	// auth token will be taken from $SFX_API_TOKEN if it exists
	// for this example, it must be set correctly
	reporter := NewReporter(NewConfig(), sfxproto.Dimensions{
		"test_dimension0": "value0",
		"test_dimension1": "value1",
	})

	gval := 0
	gauge := reporter.NewGauge("TestGauge", Value(&gval), sfxproto.Dimensions{
		"test_gauge_dimension0": "gauge0",
		"test_gauge_dimension1": "gauge1",
	})

	inc, _ := reporter.NewInc("TestIncrementer", sfxproto.Dimensions{
		"test_incrementer_dimension0": "incrementer0",
		"test_incrementer_dimension1": "incrementer1",
	})

	cval := int64(0)
	cumulative := reporter.NewCumulative("TestCumulative", Value(&cval), sfxproto.Dimensions{
		"test_cumulative_dimension0": "cumulative0",
		"test_cumulative_dimension1": "cumulative1",
	})

	atomic.AddInt64(&cval, 1)

	reporter.AddPreReportCallback(func() {
		// modify these values safely within this callback
		// modification of pointer values otherwise is not goroutine safe
		gval = 7
	})

	// incrementers are goroutine safe
	inc.Inc(1)
	inc.Inc(5)

	dps, err := reporter.Report(context.Background())

	fmt.Printf("Gauge: %d\n", gauge.IntValue())
	fmt.Printf("Incrementer: %d\n", inc.Value())
	fmt.Printf("Cumulative: %d\n", cumulative.IntValue())
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("DataPoints: %d\n", dps.Len())

	// Output:
	// Gauge: 7
	// Incrementer: 6
	// Cumulative: 1
	// Error: <nil>
	// DataPoints: 3
}
