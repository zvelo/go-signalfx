package signalfx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

func TestReporter(t *testing.T) {
	Convey("Testing Reporter", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`"OK"`))
		}))
		defer ts.Close()

		config := NewConfig()
		So(config, ShouldNotBeNil)

		config.URL = ts.URL

		r := NewReporter(config, nil)
		So(r, ShouldNotBeNil)

		So(r.datapoints.Len(), ShouldEqual, 0)
		So(len(r.buckets), ShouldEqual, 0)

		Convey("working with datapoints", func() {
			// getting datapoints

			bucket := r.NewBucket("bucket", nil)
			So(bucket, ShouldNotBeNil)
			So(r.datapoints.Len(), ShouldEqual, 0)
			So(len(r.buckets), ShouldEqual, 1)

			cumulative := r.NewCumulative("cumulative", Value(0), nil)
			So(cumulative, ShouldNotBeNil)
			So(r.datapoints.Len(), ShouldEqual, 1)
			So(len(r.buckets), ShouldEqual, 1)

			gauge := r.NewGauge("gauge", Value(0), nil)
			So(gauge, ShouldNotBeNil)
			So(r.datapoints.Len(), ShouldEqual, 2)
			So(len(r.buckets), ShouldEqual, 1)

			counter := r.NewCounter("counter", Value(0), nil)
			So(counter, ShouldNotBeNil)
			So(r.datapoints.Len(), ShouldEqual, 3)
			So(len(r.buckets), ShouldEqual, 1)

			// removing datapoints

			r.RemoveDataPoint(cumulative)
			So(r.datapoints.Len(), ShouldEqual, 2)
			So(len(r.buckets), ShouldEqual, 1)

			r.RemoveDataPoint(cumulative)
			So(r.datapoints.Len(), ShouldEqual, 2)
			So(len(r.buckets), ShouldEqual, 1)

			r.RemoveDataPoint(gauge, counter)
			So(r.datapoints.Len(), ShouldEqual, 0)
			So(len(r.buckets), ShouldEqual, 1)

			r.AddDataPoint(gauge)
			So(r.datapoints.Len(), ShouldEqual, 1)
			So(len(r.buckets), ShouldEqual, 1)

			r.AddDataPoint(cumulative, counter, gauge)
			So(r.datapoints.Len(), ShouldEqual, 3)
			So(len(r.buckets), ShouldEqual, 1)

			datapoints := NewDataPoints(3)
			datapoints.Add(cumulative, gauge, counter)
			So(datapoints.Len(), ShouldEqual, 3)

			r.RemoveDataPoints(datapoints)
			So(r.datapoints.Len(), ShouldEqual, 0)
			So(len(r.buckets), ShouldEqual, 1)

			r.RemoveBucket(bucket)
			So(r.datapoints.Len(), ShouldEqual, 0)
			So(len(r.buckets), ShouldEqual, 0)

			r.AddDataPoints(datapoints)
			So(r.datapoints.Len(), ShouldEqual, 3)
			So(len(r.buckets), ShouldEqual, 0)
		})

		Convey("adding datapoint callbacks", func() {
			So(r.datapoints.Len(), ShouldEqual, 0)

			addDataPointF := func(dims sfxproto.Dimensions) *DataPoints {
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
				return nil
			}

			r.AddDataPointsCallback(addDataPointF)
			r.AddDataPointsCallback(addDataPointErrF)

			dps, err := r.Report(nil)
			So(err, ShouldBeNil)
			So(dps, ShouldNotBeNil)
			So(dps.Len(), ShouldEqual, 2)
			So(r.datapoints.Len(), ShouldEqual, 0)
		})

		Convey("reporting should work", func() {
			ctx, cancelF := context.WithCancel(context.Background())

			r.NewBucket("bucket", nil)
			dps, err := r.Report(ctx)
			So(err, ShouldBeNil)
			So(dps.Len(), ShouldEqual, 3)

			cancelF()
			dps, err = r.Report(ctx)
			So(err, ShouldNotBeNil)
			So(dps, ShouldBeNil)
			So(err.Error(), ShouldEqual, "context canceled")

			ccopy := config.Clone()
			ccopy.URL = "z" + ts.URL
			tmpR := NewReporter(ccopy, nil)
			tmpR.NewBucket("bucket", nil)
			dps, err = tmpR.Report(context.Background())
			So(err, ShouldNotBeNil)
			So(dps, ShouldBeNil)
		})

		Convey("incrementers should work", func() {
			inc, dp := r.NewInc("Incrementer", nil)
			So(dp, ShouldNotBeNil)
			So(inc, ShouldNotBeNil)
			So(r.datapoints.Len(), ShouldEqual, 1)
			So(len(r.buckets), ShouldEqual, 0)

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

			inc, dp = r.NewCumulativeInc("CumulativeIncrementer", nil)
			So(inc, ShouldNotBeNil)
			So(r.datapoints.Len(), ShouldEqual, 2)
			So(len(r.buckets), ShouldEqual, 0)

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

	cval := 0
	cumulative := reporter.NewCumulative("TestCumulative", Value(&cval), sfxproto.Dimensions{
		"test_cumulative_dimension0": "cumulative0",
		"test_cumulative_dimension1": "cumulative1",
	})

	reporter.AddPreReportCallback(func() {
		// modify these values safely within this callback
		// modification of pointer values otherwise is not goroutine safe
		gval = 7
		cval = 1
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
