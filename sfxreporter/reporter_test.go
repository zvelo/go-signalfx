package sfxreporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxconfig"
)

func TestReporter(t *testing.T) {
	const authToken = "abc123"

	Convey("Testing Reporter", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`"OK"`))
		}))
		defer ts.Close()

		config := sfxconfig.New(authToken)
		So(config, ShouldNotBeNil)

		config.URL = ts.URL

		reporter := New(config, nil)
		So(reporter, ShouldNotBeNil)
	})
}
