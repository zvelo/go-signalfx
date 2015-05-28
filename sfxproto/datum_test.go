package sfxproto

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDatum(t *testing.T) {
	Convey("Testing Dataum", t, func() {
		Convey("created datum should have the correct value", func() {
			d := NewDatum(3)
			So(d, ShouldNotBeNil)
			So(d.IntValue, ShouldEqual, 3)
			So(d.DoubleValue, ShouldEqual, 0)
			So(d.StrValue, ShouldBeEmpty)

			d = NewDatum(0.1)
			So(d, ShouldNotBeNil)
			So(d.IntValue, ShouldEqual, 0)
			So(d.DoubleValue, ShouldEqual, 0.1)
			So(d.StrValue, ShouldBeEmpty)

			d = NewDatum("hi")
			So(d, ShouldNotBeNil)
			So(d.IntValue, ShouldEqual, 0)
			So(d.DoubleValue, ShouldEqual, 0)
			So(d.StrValue, ShouldEqual, "hi")
		})
	})
}
