package sfxproto

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDimension(t *testing.T) {
	Convey("Testing Dimension", t, func() {
		Convey("massage key should work", func() {
			So(massageKey("hello"), ShouldEqual, "hello")
			So(massageKey(".hello:bob1_&"), ShouldEqual, "_hello_bob1__")
		})
	})
}
