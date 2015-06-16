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

		Convey("adding datapoint callbacks", func() {
			So(reporter.datapoints.Len(), ShouldEqual, 0)

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

			reporter.AddDataPointsCallback(addDataPointF)
			reporter.AddDataPointsCallback(addDataPointErrF)

			ms, err := reporter.Report(nil)
			So(err, ShouldBeNil)
			So(ms, ShouldNotBeNil)
			So(ms.Len(), ShouldEqual, 2)
			So(reporter.datapoints.Len(), ShouldEqual, 0)
		})

		Convey("reporting should work", func() {
			ctx, cancelF := context.WithCancel(context.Background())

			reporter.NewBucket("bucket", nil)
			ms, err := reporter.Report(ctx)
			So(err, ShouldBeNil)
			So(ms.Len(), ShouldEqual, 3)

			cancelF()
			ms, err = reporter.Report(ctx)
			So(err, ShouldNotBeNil)
			So(ms, ShouldBeNil)
			So(err.Error(), ShouldEqual, "context canceled")

			c := config.Clone()
			c.URL = "z" + ts.URL
			tmpR := NewReporter(c, nil)
			tmpR.NewBucket("bucket", nil)
			ms, err = tmpR.Report(context.Background())
			So(err, ShouldNotBeNil)
			So(ms, ShouldBeNil)
		})
	})
}

func ExampleReporter() {
	// auth token will be taken from $SFX_API_TOKEN if it exists
	// for this example, it must be set correctly
	c := NewConfig()
	r := NewReporter(c, map[string]string{
		"test_dimension0": "value0",
		"test_dimension1": "value1",
	})

	gval := 0
	gauge := r.NewGauge("TestGauge", Value(&gval), map[string]string{
		"test_gauge_dimension0": "gauge0",
		"test_gauge_dimension1": "gauge1",
	})

	inc := r.NewInc("TestIncrementer", map[string]string{
		"test_incrementer_dimension0": "incrementer0",
		"test_incrementer_dimension1": "incrementer1",
	})

	cval := 0
	cumulative := r.NewCumulative("TestCumulative", Value(&cval), map[string]string{
		"test_cumulative_dimension0": "cumulative0",
		"test_cumulative_dimension1": "cumulative1",
	})

	gval = 7
	inc.Inc(1)
	inc.Inc(5)
	cval = 1

	_, err := r.Report(context.Background())

	fmt.Printf("gauge: %d\n", gauge.IntValue())
	fmt.Printf("incrementer: %d\n", inc.Value())
	fmt.Printf("cumulative: %d\n", cumulative.IntValue())
	fmt.Printf("error: %v\n", err)

	// Output:
	// gauge: 7
	// incrementer: 6
	// cumulative: 1
	// error: <nil>
}
