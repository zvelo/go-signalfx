package signalfx

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBackgrounding(t *testing.T) {
	Convey("Testing background reporting", t, func(c C) {
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

		// FIXME: it should be easier to override a client's transport…
		tw := transportWrapper{wrapped: reporter.client.tr}
		reporter.client.tr = &tw
		reporter.client.client = &http.Client{Transport: &tw}

		So(tw.counter, ShouldBeZeroValue)
		var count int
		counter := reporter.NewCounter("count", Value(count), nil)
		counter.Set(1)
		_, err := reporter.Report(nil)
		So(err, ShouldBeNil)
		So(tw.counter, ShouldEqual, 1)

		rj := Background(time.Second*5, func() {
			if _, err := reporter.Report(nil); err != nil {
				log.Printf("[ERR] background reporting: %s", err)
			}
		})
		counter.Set(1)
		time.Sleep(time.Second * 7)
		err = rj.Stop()
		So(err, ShouldBeNil)
		So(tw.counter, ShouldEqual, 2)
		time.Sleep(time.Second * 7)
		So(tw.counter, ShouldEqual, 2)
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

func TestBackgroundJob(t *testing.T) {
	Convey("Background jobs shoudl work", t, func() {

		var x int
		job := Background(time.Second*2, func() {
			x++
		})
		time.Sleep(time.Second * 3)
		job.Stop()
		So(x, ShouldEqual, 1)
		job.Stop()
	})
}
