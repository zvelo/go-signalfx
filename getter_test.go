package signalfx

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetter(t *testing.T) {
	Convey("Testing Getter", t, func() {
		var g Getter
		var v interface{}
		var err error

		g = GetterFunc(func() (interface{}, error) {
			return 5, nil
		})

		v, err = g.Get()
		So(v, ShouldEqual, 5)
		So(err, ShouldBeNil)

		g = Value(7)
		v, err = g.Get()
		So(v, ShouldEqual, 7)
		So(err, ShouldBeNil)

		/************************** Int32 **************************/

		i32 := NewInt32(9)
		So(i32.Value(), ShouldEqual, 9)
		g = i32
		v, err = g.Get()
		So(v, ShouldEqual, 9)
		So(err, ShouldBeNil)

		i32.Set(11)
		So(i32.Value(), ShouldEqual, 11)
		v, err = g.Get()
		So(v, ShouldEqual, 11)
		So(err, ShouldBeNil)

		i32.Inc(2)
		So(i32.Value(), ShouldEqual, 13)
		v, err = g.Get()
		So(v, ShouldEqual, 13)
		So(err, ShouldBeNil)

		/************************** Int64 **************************/

		i64 := NewInt64(9)
		So(i64.Value(), ShouldEqual, 9)
		g = i64
		v, err = g.Get()
		So(v, ShouldEqual, 9)
		So(err, ShouldBeNil)

		i64.Set(11)
		So(i64.Value(), ShouldEqual, 11)
		v, err = g.Get()
		So(v, ShouldEqual, 11)
		So(err, ShouldBeNil)

		i64.Inc(2)
		So(i64.Value(), ShouldEqual, 13)
		v, err = g.Get()
		So(v, ShouldEqual, 13)
		So(err, ShouldBeNil)

		/************************* Uint32 **************************/

		ui32 := NewUint32(9)
		So(ui32.Value(), ShouldEqual, 9)
		g = ui32
		v, err = g.Get()
		So(v, ShouldEqual, 9)
		So(err, ShouldBeNil)

		ui32.Set(11)
		So(ui32.Value(), ShouldEqual, 11)
		v, err = g.Get()
		So(v, ShouldEqual, 11)
		So(err, ShouldBeNil)

		ui32.Inc(2)
		So(ui32.Value(), ShouldEqual, 13)
		v, err = g.Get()
		So(v, ShouldEqual, 13)
		So(err, ShouldBeNil)

		/************************* Uint64 **************************/

		ui64 := NewUint64(9)
		So(ui64.Value(), ShouldEqual, 9)
		g = ui64
		v, err = g.Get()
		So(v, ShouldEqual, 9)
		So(err, ShouldBeNil)

		ui64.Set(11)
		So(ui64.Value(), ShouldEqual, 11)
		v, err = g.Get()
		So(v, ShouldEqual, 11)
		So(err, ShouldBeNil)

		ui64.Inc(2)
		So(ui64.Value(), ShouldEqual, 13)
		v, err = g.Get()
		So(v, ShouldEqual, 13)
		So(err, ShouldBeNil)
	})
}
