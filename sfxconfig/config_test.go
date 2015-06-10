package sfxconfig

import (
	"crypto/tls"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Testing Config", t, func() {
		c := New("")

		Convey("default values should be correct", func() {
			So(DefaultMaxIdleConnections, ShouldEqual, 2)
			So(DefaultTimeoutDuration, ShouldEqual, 60*time.Second)
			So(DefaultURL, ShouldEqual, "https://ingest.signalfx.com/v2/datapoint")
			So(DefaultUserAgent, ShouldEqual, "go-signalfx/"+ClientVersion)
		})

		Convey("config should be created with default values", func() {
			So(c.MaxIdleConnections, ShouldEqual, DefaultMaxIdleConnections)
			So(c.TimeoutDuration, ShouldEqual, DefaultTimeoutDuration)
			So(c.URL, ShouldEqual, DefaultURL)
			So(c.UserAgent, ShouldEqual, DefaultUserAgent)
		})

		Convey("transport should be properly configured", func() {
			tr := c.Transport()
			So(tr.TLSClientConfig, ShouldResemble, &tls.Config{InsecureSkipVerify: false})
			So(tr.MaxIdleConnsPerHost, ShouldEqual, DefaultMaxIdleConnections)
			So(tr.ResponseHeaderTimeout, ShouldEqual, DefaultTimeoutDuration)
		})
	})
}
