package signalfx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBackgroundReporter(t *testing.T) {
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

		reporter := NewBackgroundReporter(config, nil, time.Second*5)
		So(reporter, ShouldNotBeNil)

		So(reporter.datapoints.Len(), ShouldEqual, 0)
		So(len(reporter.buckets), ShouldEqual, 0)

		testReporter(config, &reporter.Reporter, ts)

		Convey("once started, a test reporter should report on intervals", func() {
			// FIXME: it should be easier to override a client's transport…
			tw := transportWrapper{wrapped: reporter.Reporter.client.tr}
			reporter.Reporter.client.tr = &tw
			reporter.Reporter.client.client = &http.Client{Transport: &tw}

			So(tw.counter, ShouldBeZeroValue)
			var count int
			_ = reporter.NewCounter("count", Value(count), nil)
			_, err := reporter.Report(nil)
			So(err, ShouldBeNil)
			So(tw.counter, ShouldEqual, 1)

			reporter.Start()
			time.Sleep(time.Second * 7)
			reporter.Stop()
			So(tw.counter, ShouldEqual, 2)
			time.Sleep(time.Second * 7)
		})
	})
}

type transportWrapper struct {
	wrapped http.RoundTripper
	counter int
}

func (tw *transportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	tw.counter++
	return tw.wrapped.RoundTrip(req)
}
