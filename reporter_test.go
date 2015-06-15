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

		So(reporter.metrics.Len(), ShouldEqual, 0)
		So(len(reporter.buckets), ShouldEqual, 0)

		Convey("working with metrics", func() {
			// getting metrics

			bucket := reporter.NewBucket("bucket", nil)
			So(bucket, ShouldNotBeNil)
			So(reporter.metrics.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			cumulative := reporter.NewCumulative("cumulative", ValueGetter(0), nil)
			So(cumulative, ShouldNotBeNil)
			So(reporter.metrics.Len(), ShouldEqual, 1)
			So(len(reporter.buckets), ShouldEqual, 1)

			gauge := reporter.NewGauge("gauge", ValueGetter(0), nil)
			So(gauge, ShouldNotBeNil)
			So(reporter.metrics.Len(), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			counter := reporter.NewCounter("counter", ValueGetter(0), nil)
			So(counter, ShouldNotBeNil)
			So(reporter.metrics.Len(), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 1)

			// removing metrics

			reporter.RemoveMetric(cumulative)
			So(reporter.metrics.Len(), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.RemoveMetric(cumulative)
			So(reporter.metrics.Len(), ShouldEqual, 2)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.RemoveMetric(gauge, counter)
			So(reporter.metrics.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.AddMetric(gauge)
			So(reporter.metrics.Len(), ShouldEqual, 1)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.AddMetric(cumulative, counter, gauge)
			So(reporter.metrics.Len(), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 1)

			metrics := NewMetrics(3)
			metrics.Add(cumulative, gauge, counter)
			So(metrics.Len(), ShouldEqual, 3)

			reporter.RemoveMetrics(metrics)
			So(reporter.metrics.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 1)

			reporter.RemoveBucket(bucket)
			So(reporter.metrics.Len(), ShouldEqual, 0)
			So(len(reporter.buckets), ShouldEqual, 0)

			reporter.AddMetrics(metrics)
			So(reporter.metrics.Len(), ShouldEqual, 3)
			So(len(reporter.buckets), ShouldEqual, 0)
		})

		Convey("adding datapoint callbacks", func() {
			So(reporter.metrics.Len(), ShouldEqual, 0)

			addMetricF := func(dims sfxproto.Dimensions) *Metrics {
				count0, err := NewCounter("count0", ValueGetter(0), nil)
				if err != nil {
					return nil
				}
				So(count0, ShouldNotBeNil)

				count1, err := NewCounter("count1", ValueGetter(0), nil)
				if err != nil {
					return nil
				}

				So(count1, ShouldNotBeNil)
				return NewMetrics(2).
					Add(count0).
					Add(count1)
			}

			addMetricErrF := func(dims sfxproto.Dimensions) *Metrics {
				return nil
			}

			reporter.AddMetricsCallback(addMetricF)
			reporter.AddMetricsCallback(addMetricErrF)

			ms, err := reporter.Report(nil)
			So(err, ShouldBeNil)
			So(ms, ShouldNotBeNil)
			So(ms.Len(), ShouldEqual, 2)
			So(reporter.metrics.Len(), ShouldEqual, 0)
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
	config := NewConfig()
	reporter := NewReporter(config, map[string]string{
		"test_dimension0": "value0",
		"test_dimension1": "value1",
	})

	inc := reporter.NewIncrementer("TestIncrementer", map[string]string{
		"test_counter_dimension0": "counter0",
		"test_counter_dimension1": "counter1",
	})

	cval := 0
	cumulative := reporter.NewCumulative("TestCumulative", ValueGetter(&cval), map[string]string{
		"test_cumulative_dimension0": "cumulative0",
		"test_cumulative_dimension1": "cumulative1",
	})

	inc.Inc(1)
	inc.Inc(5)
	cval = 1

	_, err := reporter.Report(context.Background())

	fmt.Printf("incrementer: %d\n", inc.Value())
	fmt.Printf("cumulative: %d\n", cumulative.IntValue())
	fmt.Printf("error: %v\n", err)

	// Output:
	// incrementer: 6
	// cumulative: 1
	// error: <nil>
}
