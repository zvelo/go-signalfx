package signalfx

import (
	"errors"
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

		Convey("report should fail when Getters return an error", func() {
			ccopy := config.Clone()
			ccopy.URL = "z" + ts.URL
			tmpR := NewReporter(ccopy, nil)

			// when adding the getter, its value is taken, and that has to not
			// return an error. then after that first Get, it should return an
			// error for this test to work
			i := 0
			f := GetterFunc(func() (interface{}, error) {
				switch i {
				case 0:
					i++
					return 0, nil
				default:
					return nil, errors.New("bad getter")
				}
			})

			dp := tmpR.NewGauge("BadGauge", f, nil)
			So(dp, ShouldNotBeNil)
			dps, err := tmpR.Report(context.Background())
			So(dps, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "bad getter")
		})

		Convey("Getters should work", func() {
			i32, dpi32 := reporter.NewInt32("Int32", nil)
			ci32, dpci32 := reporter.NewCumulativeInt32("CumulativeInt32", nil)
			i64, dpi64 := reporter.NewInt64("Int64", nil)
			ci64, dpci64 := reporter.NewCumulativeInt64("CumulativeInt64", nil)
			ui32, dpui32 := reporter.NewUint32("Uint32", nil)
			cui32, dpcui32 := reporter.NewCumulativeUint32("CumulativeUint32", nil)
			ui64, dpui64 := reporter.NewUint64("Uint64", nil)
			cui64, dpcui64 := reporter.NewCumulativeUint64("CumulativeUint64", nil)

			So(reporter.datapoints.Len(), ShouldEqual, 8)
			So(len(reporter.buckets), ShouldEqual, 0)

			Convey("Metric names should be required", func() {
				i32, dpi32 := reporter.NewInt32("", nil)
				So(i32, ShouldBeNil)
				So(dpi32, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)

				ci32, dpci32 := reporter.NewCumulativeInt32("", nil)
				So(ci32, ShouldBeNil)
				So(dpci32, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)

				i64, dpi64 := reporter.NewInt64("", nil)
				So(i64, ShouldBeNil)
				So(dpi64, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)

				ci64, dpci64 := reporter.NewCumulativeInt64("", nil)
				So(ci64, ShouldBeNil)
				So(dpci64, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)

				ui32, dpui32 := reporter.NewUint32("", nil)
				So(ui32, ShouldBeNil)
				So(dpui32, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)

				cui32, dpcui32 := reporter.NewCumulativeUint32("", nil)
				So(cui32, ShouldBeNil)
				So(dpcui32, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)

				ui64, dpui64 := reporter.NewUint64("", nil)
				So(ui64, ShouldBeNil)
				So(dpui64, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)

				cui64, dpcui64 := reporter.NewCumulativeUint64("", nil)
				So(cui64, ShouldBeNil)
				So(dpcui64, ShouldBeNil)
				So(reporter.datapoints.Len(), ShouldEqual, 8)
				So(len(reporter.buckets), ShouldEqual, 0)
			})

			Convey("Int32", func() {
				So(i32, ShouldNotBeNil)
				So(dpi32, ShouldNotBeNil)
				So(i32.Value(), ShouldEqual, 0)
				v, err := i32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(i32.Value(), ShouldEqual, dpi32.IntValue())

				i32.Set(5)
				So(i32.Value(), ShouldEqual, 5)
				v, err = i32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(i32.Value(), ShouldEqual, dpi32.IntValue())

				i32.Inc(1)
				So(i32.Value(), ShouldEqual, 6)
				v, err = i32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(i32.Value(), ShouldEqual, dpi32.IntValue())
			})

			Convey("CumulativeInt32", func() {
				So(ci32, ShouldNotBeNil)
				So(dpci32, ShouldNotBeNil)
				So(ci32.Value(), ShouldEqual, 0)
				v, err := ci32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(ci32.Value(), ShouldEqual, dpci32.IntValue())

				ci32.Set(5)
				So(ci32.Value(), ShouldEqual, 5)
				v, err = ci32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(ci32.Value(), ShouldEqual, dpci32.IntValue())

				ci32.Inc(1)
				So(ci32.Value(), ShouldEqual, 6)
				v, err = ci32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(ci32.Value(), ShouldEqual, dpci32.IntValue())
			})

			Convey("Int64", func() {
				So(i64, ShouldNotBeNil)
				So(dpi64, ShouldNotBeNil)
				So(i64.Value(), ShouldEqual, 0)
				v, err := i64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(i64.Value(), ShouldEqual, dpi64.IntValue())

				i64.Set(5)
				So(i64.Value(), ShouldEqual, 5)
				v, err = i64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(i64.Value(), ShouldEqual, dpi64.IntValue())

				i64.Inc(1)
				So(i64.Value(), ShouldEqual, 6)
				v, err = i64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(i64.Value(), ShouldEqual, dpi64.IntValue())
			})

			Convey("CumulativeInt64", func() {
				So(ci64, ShouldNotBeNil)
				So(dpci64, ShouldNotBeNil)
				So(ci64.Value(), ShouldEqual, 0)
				v, err := ci64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(ci64.Value(), ShouldEqual, dpci64.IntValue())

				ci64.Set(5)
				So(ci64.Value(), ShouldEqual, 5)
				v, err = ci64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(ci64.Value(), ShouldEqual, dpci64.IntValue())

				ci64.Inc(1)
				So(ci64.Value(), ShouldEqual, 6)
				v, err = ci64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(ci64.Value(), ShouldEqual, dpci64.IntValue())
			})

			Convey("Uint32", func() {
				So(ui32, ShouldNotBeNil)
				So(dpui32, ShouldNotBeNil)
				So(ui32.Value(), ShouldEqual, 0)
				v, err := ui32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(ui32.Value(), ShouldEqual, dpui32.IntValue())

				ui32.Set(5)
				So(ui32.Value(), ShouldEqual, 5)
				v, err = ui32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(ui32.Value(), ShouldEqual, dpui32.IntValue())

				ui32.Inc(1)
				So(ui32.Value(), ShouldEqual, 6)
				v, err = ui32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(ui32.Value(), ShouldEqual, dpui32.IntValue())
			})

			Convey("CumulativeUint32", func() {
				So(cui32, ShouldNotBeNil)
				So(dpcui32, ShouldNotBeNil)
				So(cui32.Value(), ShouldEqual, 0)
				v, err := cui32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(cui32.Value(), ShouldEqual, dpcui32.IntValue())

				cui32.Set(5)
				So(cui32.Value(), ShouldEqual, 5)
				v, err = cui32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(cui32.Value(), ShouldEqual, dpcui32.IntValue())

				cui32.Inc(1)
				So(cui32.Value(), ShouldEqual, 6)
				v, err = cui32.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(cui32.Value(), ShouldEqual, dpcui32.IntValue())
			})

			Convey("Uint64", func() {
				So(ui64, ShouldNotBeNil)
				So(dpui64, ShouldNotBeNil)
				So(ui64.Value(), ShouldEqual, 0)
				v, err := ui64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(ui64.Value(), ShouldEqual, dpui64.IntValue())

				ui64.Set(5)
				So(ui64.Value(), ShouldEqual, 5)
				v, err = ui64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(ui64.Value(), ShouldEqual, dpui64.IntValue())

				ui64.Inc(1)
				So(ui64.Value(), ShouldEqual, 6)
				v, err = ui64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(ui64.Value(), ShouldEqual, dpui64.IntValue())
			})

			Convey("CumulativeUint64", func() {
				So(cui64, ShouldNotBeNil)
				So(dpcui64, ShouldNotBeNil)
				So(cui64.Value(), ShouldEqual, 0)
				v, err := cui64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 0)
				So(cui64.Value(), ShouldEqual, dpcui64.IntValue())

				cui64.Set(5)
				So(cui64.Value(), ShouldEqual, 5)
				v, err = cui64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 5)
				So(cui64.Value(), ShouldEqual, dpcui64.IntValue())

				cui64.Inc(1)
				So(cui64.Value(), ShouldEqual, 6)
				v, err = cui64.Get()
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 6)
				So(cui64.Value(), ShouldEqual, dpcui64.IntValue())
			})
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

	i, _ := reporter.NewInt64("TestInt64", sfxproto.Dimensions{
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
	i.Inc(1)
	i.Inc(5)

	dps, err := reporter.Report(context.Background())

	fmt.Printf("Gauge: %d\n", gauge.IntValue())
	fmt.Printf("Incrementer: %d\n", i.Value())
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
