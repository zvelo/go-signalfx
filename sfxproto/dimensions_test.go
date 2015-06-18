package sfxproto

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

func toMap(in []*Dimension) map[string]string {
	ret := make(map[string]string, len(in))
	for _, dim := range in {
		ret[*dim.Key] = *dim.Value
	}
	return ret
}

func TestDimensions(t *testing.T) {
	Convey("Testing Dimensions", t, func() {
		dims := Dimensions{
			"one": "1",
			"":    "",
			"two": "2",
		}
		So(len(dims), ShouldEqual, 3)

		dslice := []*Dimension{
			{Key: proto.String("one"), Value: proto.String("1")},
			{Key: proto.String("two"), Value: proto.String("2")},
		}

		Convey("List should work", func() {
			So(toMap(dims.List()), ShouldResemble, toMap(dslice))
		})

		Convey("Append should work", func() {
			dims2 := Dimensions{"three": "3"}
			So(len(dims2), ShouldEqual, 1)

			tmp := dims.Append(dims2)
			So(len(dims), ShouldEqual, 3)
			So(len(tmp), ShouldEqual, 4)

			So(tmp, ShouldResemble, Dimensions{
				"one":   "1",
				"":      "",
				"two":   "2",
				"three": "3",
			})
		})

		Convey("massage key should work", func() {
			So(massageKey("hello"), ShouldEqual, "hello")
			So(massageKey(".hello:bob1_&"), ShouldEqual, "_hello_bob1__")
		})

		Convey("Clone should work", func() {
			dims2 := dims.Clone()
			So(dims, ShouldNotEqual, dims2)
			So(dims, ShouldResemble, dims2)
		})

		Convey("NewDimensions should work", func() {
			dims2 := NewDimensions(dslice)
			So(dims2, ShouldResemble, Dimensions{
				"one": "1",
				"two": "2",
			})

			dims2 = NewDimensions(nil)
			So(dims2, ShouldResemble, Dimensions{})
		})
	})
}
