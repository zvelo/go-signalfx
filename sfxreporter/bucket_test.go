package sfxreporter

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxproto"
)

func TestBucket(t *testing.T) {
	Convey("Testing Bucket", t, func() {
		b := NewBucket("test", sfxproto.Dimensions{"c": "c"})
		So(b.dataPoints(nil).Len(), ShouldEqual, 3)

		b.Add(1)
		b.Add(2)
		b.Add(3)

		So(b.Max(), ShouldEqual, 3)
		So(b.Min(), ShouldEqual, 1)
		So(b.Count(), ShouldEqual, 3)
		So(b.Sum(), ShouldEqual, 6)
		So(b.SumOfSquares(), ShouldEqual, 14)

		b.Add(4)
		b.Add(10)
		b.Add(1)

		So(b.Max(), ShouldEqual, 10)
		So(b.Min(), ShouldEqual, 1)

		b.Add(1)
		So(b.dataPoints(sfxproto.Dimensions{"a": "b"}).Len(), ShouldEqual, 5)
	})
}
