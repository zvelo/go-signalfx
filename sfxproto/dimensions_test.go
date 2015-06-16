package sfxproto

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDimensions(t *testing.T) {
	Convey("Testing Dimensions", t, func() {
		Convey("massage key should work", func() {
			So(massageKey("hello"), ShouldEqual, "hello")
			So(massageKey(".hello:bob1_&"), ShouldEqual, "_hello_bob1__")
		})
	})
}
