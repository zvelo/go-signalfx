package signalfx

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxproto"
)

func TestBucket(t *testing.T) {
	Convey("Testing Bucket", t, func() {
		b := NewBucket("test", sfxproto.Dimensions{"c": "3"})
		So(b, ShouldNotBeNil)
		So(b.DataPoints(nil).Len(), ShouldEqual, 3)

		Convey("metric naming should be correct", func() {
			So(b.Metric(), ShouldEqual, "test")
			b.SetMetric("new metric name")
			So(b.Metric(), ShouldEqual, "new metric name")
		})

		Convey("dimension operations should work", func() {
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{"c": "3"})

			b.SetDimension("a", "")
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{
				"c": "3",
			})

			b.SetDimension("", "1")
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{
				"c": "3",
			})

			b.SetDimension("", "")
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{
				"c": "3",
			})

			b.SetDimension("a", "1")
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{
				"a": "1",
				"c": "3",
			})

			dims := sfxproto.Dimensions{"b": "2"}
			b.SetDimensions(dims)
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{
				"a": "1",
				"b": "2",
				"c": "3",
			})

			b.RemoveDimension("c")
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{
				"a": "1",
				"b": "2",
			})

			b.RemoveDimension("a", "b")
			So(b.Dimensions(), ShouldResemble, sfxproto.Dimensions{})
		})

		Convey("data handling should be correct", func() {
			b.Add(1)
			b.Add(2)
			b.Add(3)

			So(b.Max(), ShouldEqual, 3)
			So(b.Min(), ShouldEqual, 1)
			So(b.Count(), ShouldEqual, 3)
			So(b.Sum(), ShouldEqual, 6)
			So(b.SumOfSquares(), ShouldEqual, 14)

			b.Add(0)
			b.Add(10)
			b.Add(1)

			So(b.Max(), ShouldEqual, 10)
			So(b.Min(), ShouldEqual, 0)

			b.Add(1)
			So(b.DataPoints(sfxproto.Dimensions{"a": "b"}).Len(), ShouldEqual, 5)
		})
	})
}
