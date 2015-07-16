package signalfx

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataPoints(t *testing.T) {
	Convey("Testing DataPoints", t, func() {
		dps := NewDataPoints(0)
		So(dps.Len(), ShouldEqual, 0)

		dps = NewDataPoints(2)
		So(dps.Len(), ShouldEqual, 0)

		So(dps.Add(nil), ShouldEqual, dps)
		So(dps.Len(), ShouldEqual, 0)

		i := 0
		m0, err := NewCounter("m0", GetterFunc(func() (interface{}, error) {
			i++

			// when added and the first time DataPoints is called, this should work
			if i < 3 {
				return i, nil
			}

			// but this is an invalid data type and calling DataPoints a second
			// time should fail
			return true, nil
		}), nil)

		So(err, ShouldBeNil)
		So(dps.Add(m0), ShouldEqual, dps)
		So(dps.Len(), ShouldEqual, 1)

		m1, err := NewCounter("m1", 1, nil)
		So(err, ShouldBeNil)
		So(dps.Add(m1), ShouldEqual, dps)
		So(dps.Len(), ShouldEqual, 2)

		m2, err := NewCounter("m2", 2, nil)
		So(err, ShouldBeNil)
		So(dps.Add(m2), ShouldEqual, dps)
		So(dps.Len(), ShouldEqual, 3)

		dps.Remove(nil)
		So(dps.Len(), ShouldEqual, 3)

		dps.Remove(m1)
		So(dps.Len(), ShouldEqual, 2)

		dps.Remove(m1)
		So(dps.Len(), ShouldEqual, 2)

		rs := NewDataPoints(0)
		rs.Add(m2)
		So(rs.Len(), ShouldEqual, 1)

		dps.RemoveDataPoints(rs)
		So(dps.Len(), ShouldEqual, 1)

		pdps, err := dps.ProtoDataPoints()
		So(err, ShouldBeNil)
		So(pdps.Len(), ShouldEqual, 1)
	})
}
