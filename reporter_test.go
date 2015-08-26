package signalfx

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

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

			cumulative := &CumulativeCounter{Metric: "cumulative"}
			reporter.Track(cumulative)
			So(len(reporter.metrics), ShouldEqual, 1)
			So(len(reporter.buckets), ShouldEqual, 1)

			gauge := &Gauge{Metric: "gauge"}
			reporter.Track(gauge)
			So(len(reporter.metrics), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			counter := &Counter{Metric: "counter"}
			reporter.Track(counter)
			So(counter, ShouldNotBeNil)
			So(len(reporter.metrics), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 1)

			// removing datapoints

			reporter.Untrack(cumulative)
			So(len(reporter.metrics), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.Untrack(cumulative)
			So(len(reporter.metrics), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.Untrack(gauge, counter)
			So(len(reporter.metrics), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.Track(gauge)
			So(len(reporter.metrics), ShouldEqual, 1)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.Track(cumulative, counter, gauge)
			So(len(reporter.metrics), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.Untrack(counter, gauge, cumulative)
			So(len(reporter.metrics), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.RemoveBucket(bucket)
			So(len(reporter.metrics), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 0)
		})

		Convey("callbacks should be fired", func() {
			So(reporter.datapoints.Len(), ShouldEqual, 0)

			cb := 0
			addDataPointF := func(dims map[string]string) *DataPoints {
				cb++
				count0, err := NewCounter("count0", Value(1), nil)
				if err != nil {
					return nil
				}
				So(count0, ShouldNotBeNil)

				count1, err := NewCounter("count1", Value(1), nil)
				if err != nil {
					return nil
				}

				So(count1, ShouldNotBeNil)
				return NewDataPoints(2).
					Add(count0).
					Add(count1)
			}
			addDataPointErrF := func(dims map[string]string) *DataPoints {
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
			bucket := reporter.NewBucket("bucket", nil)
			bucket.Add(2)
			dps, err := reporter.Report(context.Background())
			So(err, ShouldBeNil)
			So(dps.Len(), ShouldEqual, 5) // TODO: verify that 5 is correct, not just expected
		})

		Convey("a blanked counter shouldn't report", func() {
			counter := reporter.NewCounter("foo", Value(0), nil)
			_ = counter
			_, err := reporter.Report(context.Background())
			So(err, ShouldEqual, sfxproto.ErrMarshalNoData)
		})

		Convey("a cumulative counter shouldn't report the same value", func() {
			counter := reporter.NewCumulative("foo", Value(0), nil)
			_, err := reporter.Report(context.Background())
			So(err, ShouldEqual, sfxproto.ErrMarshalNoData)
			counter.Set(1)
			_, err = reporter.Report(context.Background())
			So(err, ShouldBeNil)
			// since it didn't change, it shouldn't report
			_, err = reporter.Report(context.Background())
			So(err, ShouldEqual, sfxproto.ErrMarshalNoData)
			counter.Set(2)
			_, err = reporter.Report(context.Background())
			So(err, ShouldBeNil)
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
			bucket := reporter.NewBucket("bucket", nil)
			bucket.Add(1)
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
			bucket := tmpR.NewBucket("bucket", nil)
			bucket.Add(1)
			dps, err := tmpR.Report(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Post z"+ts.URL+": unsupported protocol scheme \"zhttp\"")
			So(dps, ShouldBeNil)
		})

		Convey("Inc should handle cheap one-shot counter increments", func() {
			config := config.Clone()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`"OK"`))
			}))
			defer ts.Close()
			config.URL = ts.URL
			r := NewReporter(config, map[string]string{"foo": "bar"})

			// FIXME: it _really_ should be easier to
			// override a reporter's client…
			tw := transportWrapper{wrapped: r.client.tr}
			r.client.tr = &tw
			r.client.client = &http.Client{Transport: &tw}
			So(tw.counter, ShouldBeZeroValue)

			hostname, err := os.Hostname()
			if err != nil {
				hostname = fmt.Sprintf("zvelo-testing-host-%d", os.Getpid())
			}
			r.Inc("test-metric", map[string]string{
				"client": "Wilkinson et Cie",
				"host":   hostname,
				"pid":    strconv.Itoa(os.Getpid()),
			}, 1)
			_, err = r.Report(context.Background())
			So(err, ShouldBeNil)
			So(tw.counter, ShouldEqual, 1)
		})

		Convey("report does not include broken Getters", func() {
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
			So(err.Error(), ShouldEqual, "no data to marshal")
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
		Convey("Testing background reporting", func() {
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

			// FIXME: it should be easier to override a client's transport…
			tw := transportWrapper{wrapped: reporter.client.tr}
			reporter.client.tr = &tw
			reporter.client.client = &http.Client{Transport: &tw}

			So(tw.counter, ShouldBeZeroValue)
			var count int
			counter := reporter.NewCounter("count", Value(count), nil)
			err := counter.Set(1)
			So(err, ShouldBeNil)
			_, err = reporter.Report(nil)
			So(err, ShouldBeNil)
			So(tw.counter, ShouldEqual, 1)

			cancelFunc := reporter.RunInBackground(time.Second * 5)
			err = counter.Set(2)
			So(err, ShouldBeNil)
			time.Sleep(time.Second * 7)
			So(tw.counter, ShouldEqual, 2)
			// let it run once more, with no data to send
			time.Sleep(time.Second * 7)
			So(tw.counter, ShouldEqual, 2)
			cancelFunc()
			// prove that it's cancelled
			err = counter.Set(3)
			So(err, ShouldBeNil)
			time.Sleep(time.Second * 7)
			So(tw.counter, ShouldEqual, 2)
		})
	})
}

func ExampleReporter() {
	// auth token will be taken from $SFX_API_TOKEN if it exists
	// for this example, it must be set correctly
	reporter := NewReporter(NewConfig(), map[string]string{
		"test_dimension0": "value0",
		"test_dimension1": "value1",
	})

	gval := 0
	gauge := reporter.NewGauge("TestGauge", Value(&gval), map[string]string{
		"test_gauge_dimension0": "gauge0",
		"test_gauge_dimension1": "gauge1",
	})

	i, _ := reporter.NewInt64("TestInt64", map[string]string{
		"test_incrementer_dimension0": "incrementer0",
		"test_incrementer_dimension1": "incrementer1",
	})

	cval := int64(0)
	cumulative := reporter.NewCumulative("TestCumulative", Value(&cval), map[string]string{
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
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Gauge: %d\n", gauge.IntValue())
		fmt.Printf("Incrementer: %d\n", i.Value())
		fmt.Printf("Cumulative: %d\n", cumulative.IntValue())
		fmt.Printf("DataPoints: %d\n", dps.Len())
	}

	// Output:
	// Gauge: 7
	// Incrementer: 6
	// Cumulative: 1
	// DataPoints: 3
}

type transportWrapper struct {
	wrapped http.RoundTripper
	counter int
}

func (tw *transportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	tw.counter++
	return tw.wrapped.RoundTrip(req)
}
