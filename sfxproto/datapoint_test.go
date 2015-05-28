package sfxproto

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataPoint(t *testing.T) {
	Convey("Testing DataPoint", t, func() {
		Convey("massage key should work", func() {
			So(massageKey("hello"), ShouldEqual, "hello")
			So(massageKey(".hello:bob1_&"), ShouldEqual, "_hello_bob1__")
		})

		Convey("setting values should work", func() {
		})
	})
}
