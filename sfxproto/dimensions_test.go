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
		So(dims.Equal(dims), ShouldBeTrue)

		dslice := []*Dimension{
			{Key: proto.String("one"), Value: proto.String("1")},
			{Key: proto.String("two"), Value: proto.String("2")},
		}

		Convey("List should work", func() {
			So(toMap(dims.List()), ShouldResemble, toMap(dslice))
		})

		Convey("Append should work", func() {
			So(len(dims), ShouldEqual, 3)

			tst := Dimensions{
				"one":   "1",
				"":      "",
				"two":   "2",
				"three": "3",
			}

			So(dims.Equal(tst), ShouldBeFalse)
			So(tst.Equal(dims), ShouldBeFalse)
			So(tst.Equal(tst), ShouldBeTrue)

			dims2 := Dimensions{"three": "3"}
			So(len(dims2), ShouldEqual, 1)
			So(dims2.Equal(tst), ShouldBeFalse)
			So(tst.Equal(dims2), ShouldBeFalse)

			tmp := dims.Append(dims2)
			So(len(tmp), ShouldEqual, 4)
			So(tmp, ShouldResemble, tst)
			So(tmp.Equal(tst), ShouldBeTrue)
			So(tst.Equal(tmp), ShouldBeTrue)
		})

		Convey("massage key should work", func() {
			So(massageKey("hello"), ShouldEqual, "hello")
			So(massageKey(".hello:bob1_&"), ShouldEqual, "_hello_bob1__")
		})

		Convey("Clone/Equal should work", func() {
			dims2 := dims.Clone()
			So(dims, ShouldNotEqual, dims2)
			So(dims, ShouldResemble, dims2)
			So(dims.Equal(dims2), ShouldBeTrue)
			So(dims2.Equal(dims), ShouldBeTrue)

			dims2 = dims.Clone()
			dims2["one"] = "one"
			So(dims.Equal(dims2), ShouldBeFalse)
			So(dims2.Equal(dims), ShouldBeFalse)

			dims2 = Dimensions{}
			So(dims.Equal(dims2), ShouldBeFalse)
			So(dims2.Equal(dims), ShouldBeFalse)

			dims2["a"] = "1"
			dims2["b"] = "2"
			dims2["c"] = "3"

			So(dims.Equal(dims2), ShouldBeFalse)
			So(dims2.Equal(dims), ShouldBeFalse)

			dims2 = dims.Clone()
			dims2["one"] = "one"
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
