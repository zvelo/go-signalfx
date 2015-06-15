package signalfx

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Testing Config", t, func() {
		c := NewConfig()

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
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`"OK"`))
			}))
			defer ts.Close()

			c.URL = ts.URL
			tr := c.Transport()
			So(tr.TLSClientConfig, ShouldResemble, &tls.Config{InsecureSkipVerify: false})
			So(tr.MaxIdleConnsPerHost, ShouldEqual, DefaultMaxIdleConnections)
			So(tr.ResponseHeaderTimeout, ShouldEqual, DefaultTimeoutDuration)

			u, err := url.Parse(c.URL)
			So(err, ShouldBeNil)

			conn, err := tr.Dial("tcp", u.Host)
			So(err, ShouldBeNil)
			So(conn, ShouldNotBeNil)
			So(conn.Close(), ShouldBeNil)
		})

		Convey("clone should work", func() {
			c0 := NewConfig()

			if envToken := os.Getenv("SFX_API_TOKEN"); len(envToken) > 0 {
				So(c0.AuthToken, ShouldEqual, envToken)
			}

			c1 := c0.Clone()
			So(c0, ShouldNotEqual, c1)
			So(c0.MaxIdleConnections, ShouldEqual, c1.MaxIdleConnections)
			So(c0.TimeoutDuration, ShouldEqual, c1.TimeoutDuration)
			So(c0.URL, ShouldEqual, c1.URL)
			So(c0.AuthToken, ShouldEqual, c1.AuthToken)
			So(c0.UserAgent, ShouldEqual, c1.UserAgent)
			So(c0.TLSInsecureSkipVerify, ShouldEqual, c1.TLSInsecureSkipVerify)
		})
	})
}
