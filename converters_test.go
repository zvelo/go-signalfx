package sfxmetric

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConverters(t *testing.T) {
	Convey("Testing Converters", t, func() {
		Convey("toInt64", func() {
			Convey("int", func() {
				val := int(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("int8", func() {
				val := int8(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("int16", func() {
				val := int16(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("int32", func() {
				val := int32(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("int64", func() {
				val := int64(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("uint", func() {
				val := uint(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("uint8", func() {
				val := uint8(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("uint16", func() {
				val := uint16(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("uint32", func() {
				val := uint32(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("uint64", func() {
				val := uint64(5)
				i, err := toInt64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toInt64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("non-int types", func() {
				f32 := float32(5)
				_, err := toInt64(f32)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toInt64(&f32)
				So(err, ShouldEqual, ErrIllegalType)

				f64 := float64(5)
				_, err = toInt64(f64)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toInt64(&f64)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toInt64(true)
				So(err, ShouldEqual, ErrIllegalType)

				str := ""
				_, err = toInt64(str)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toInt64(&str)
				So(err, ShouldEqual, ErrIllegalType)
			})
		})

		Convey("toFloat64", func() {
			Convey("float32", func() {
				val := float32(5)
				i, err := toFloat64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toFloat64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("float64", func() {
				val := float64(5)
				i, err := toFloat64(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)

				i, err = toFloat64(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, 5)
			})

			Convey("non-float types", func() {
				i := int(5)
				_, err := toFloat64(i)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toFloat64(&i)
				So(err, ShouldEqual, ErrIllegalType)

				u32 := uint32(5)
				_, err = toFloat64(u32)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toFloat64(&u32)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toFloat64(true)
				So(err, ShouldEqual, ErrIllegalType)

				str := ""
				_, err = toFloat64(str)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toFloat64(&str)
				So(err, ShouldEqual, ErrIllegalType)
			})
		})

		Convey("toString", func() {
			Convey("valid string", func() {
				val := "abc"
				i, err := toString(val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, "abc")

				i, err = toString(&val)
				So(err, ShouldBeNil)
				So(i, ShouldEqual, "abc")
			})

			Convey("non-string types", func() {
				i := int(5)
				_, err := toString(i)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toString(&i)
				So(err, ShouldEqual, ErrIllegalType)

				u32 := uint32(5)
				_, err = toString(u32)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toString(&u32)
				So(err, ShouldEqual, ErrIllegalType)

				f32 := float32(5)
				_, err = toString(f32)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toString(&f32)
				So(err, ShouldEqual, ErrIllegalType)

				f64 := float64(5)
				_, err = toString(f64)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toString(&f64)
				So(err, ShouldEqual, ErrIllegalType)

				_, err = toString(true)
				So(err, ShouldEqual, ErrIllegalType)
			})
		})
	})
}
