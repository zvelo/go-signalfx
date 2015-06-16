package signalfx

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataPoints(t *testing.T) {
	Convey("Testing DataPoints", t, func() {
		ms := NewDataPoints(0)
		So(ms.Len(), ShouldEqual, 0)

		ms = NewDataPoints(2)
		So(ms.Len(), ShouldEqual, 0)

		So(ms.Add(nil), ShouldEqual, ms)
		So(ms.Len(), ShouldEqual, 0)

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
		So(ms.Add(m0), ShouldEqual, ms)
		So(ms.Len(), ShouldEqual, 1)

		m1, err := NewCounter("m1", 1, nil)
		So(err, ShouldBeNil)
		So(ms.Add(m1), ShouldEqual, ms)
		So(ms.Len(), ShouldEqual, 2)

		m2, err := NewCounter("m2", 2, nil)
		So(err, ShouldBeNil)
		So(ms.Add(m2), ShouldEqual, ms)
		So(ms.Len(), ShouldEqual, 3)

		ms.Remove(nil)
		So(ms.Len(), ShouldEqual, 3)

		ms.Remove(m1)
		So(ms.Len(), ShouldEqual, 2)

		ms.Remove(m1)
		So(ms.Len(), ShouldEqual, 2)

		rs := NewDataPoints(0)
		rs.Add(m2)
		So(rs.Len(), ShouldEqual, 1)

		ms.RemoveDataPoints(rs)
		So(ms.Len(), ShouldEqual, 1)

		dps, err := ms.DataPoints()
		So(err, ShouldBeNil)
		So(dps.Len(), ShouldEqual, 1)

		dps, err = ms.DataPoints()
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrIllegalType)
		So(dps, ShouldBeNil)
	})
}
