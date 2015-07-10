package signalfx

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxproto"
)

func TestDataPoint(t *testing.T) {
	Convey("Testing DataPoint", t, func() {
		Convey("basic functionality", func() {
			c, err := NewCounter("count", 2, nil)
			So(err, ShouldBeNil)
			So(c.IntValue(), ShouldEqual, 2)
			So(c.Metric(), ShouldEqual, "count")
			So(c.Time().Before(time.Now()), ShouldBeTrue)
			So(c.Type(), ShouldEqual, sfxproto.MetricType_COUNTER)

			dim := c.Dimensions()
			So(dim, ShouldNotBeNil)
			So(len(dim), ShouldEqual, 0)

			e := c.Clone()
			So(e, ShouldNotEqual, c)
			So(e.pdp, ShouldNotEqual, c.pdp)
			So(e.IntValue(), ShouldEqual, 2)
			So(e.Metric(), ShouldEqual, "count")
			So(e.Time().Equal(c.Time()), ShouldBeTrue)
			So(e.Type(), ShouldEqual, sfxproto.MetricType_COUNTER)

			So(c.Equal(e), ShouldBeTrue)
			So(e.Equal(c), ShouldBeTrue)

			e.SetMetric("another_metric")
			So(e.Metric(), ShouldEqual, "another_metric")
		})

		Convey("created datapoint should have the correct value", func() {
			m0, err := NewCounter("count", 3, nil)
			So(err, ShouldBeNil)
			So(m0, ShouldNotBeNil)
			So(m0.IntValue(), ShouldEqual, 3)
			So(m0.DoubleValue(), ShouldEqual, 0)
			So(m0.StrValue(), ShouldBeEmpty)

			m1, err := NewCounter("count", 0.1, nil)
			So(err, ShouldBeNil)
			So(m1, ShouldNotBeNil)
			So(m1.IntValue(), ShouldEqual, 0)
			So(m1.DoubleValue(), ShouldEqual, 0.1)
			So(m1.StrValue(), ShouldBeEmpty)
			So(m0.Equal(m1), ShouldBeFalse)
			So(m1.Equal(m0), ShouldBeFalse)

			m2, err := NewCounter("count", "hi", nil)
			So(err, ShouldBeNil)
			So(m2, ShouldNotBeNil)
			So(m2.IntValue(), ShouldEqual, 0)
			So(m2.DoubleValue(), ShouldEqual, 0)
			So(m2.StrValue(), ShouldEqual, "hi")

			err = m2.Set(nil)
			So(err, ShouldBeNil)
			So(m2.IntValue(), ShouldEqual, 0)
			So(m2.DoubleValue(), ShouldEqual, 0)
			So(m2.StrValue(), ShouldEqual, "")

			err = m2.update()
			So(err, ShouldBeNil)
		})

		Convey("Value functionality", func() {
			i := 5
			g := Value(&i)
			v, err := g.Get()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, &i)

			c, err := NewCumulative("count", g, nil)
			So(err, ShouldBeNil)
			So(c.IntValue(), ShouldEqual, 5)
			So(c.Type(), ShouldEqual, sfxproto.MetricType_CUMULATIVE_COUNTER)

			i = 9
			So(c.update(), ShouldBeNil)
			So(c.IntValue(), ShouldEqual, 9)
		})

		Convey("GetterFunc functionality", func() {
			i := 9
			c, err := NewGauge("count", GetterFunc(func() (interface{}, error) { return i, nil }), nil)
			So(err, ShouldBeNil)
			So(c.IntValue(), ShouldEqual, 9)
			So(c.Type(), ShouldEqual, sfxproto.MetricType_GAUGE)

			i++
			So(c.update(), ShouldBeNil)
			So(c.IntValue(), ShouldEqual, 10)
		})

		Convey("illegal types should be rejected", func() {
			c, err := NewCounter("count", false, nil)
			So(c, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrIllegalType)
		})

		Convey("GetterFunc functionality with illegal type", func() {
			c, err := NewCounter("count", GetterFunc(func() (interface{}, error) { return true, nil }), nil)
			So(c, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrIllegalType)
		})

		Convey("GetterFunc functionality with illegal type change", func() {
			var val interface{}
			val = 8
			c, err := NewCounter("count", GetterFunc(func() (interface{}, error) { return val, nil }), nil)
			So(c, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(c.IntValue(), ShouldEqual, 8)

			val = true
			err = c.update()
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrIllegalType)
		})

		Convey("GetterFunc functionality with changed return type", func() {
			i := 0
			c, err := NewCounter("count", GetterFunc(func() (interface{}, error) {
				if i == 0 {
					i++
					return i, nil
				}

				return i, fmt.Errorf("oh noes")
			}), nil)
			So(c, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(c.IntValue(), ShouldEqual, 0)

			err = c.update()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "oh noes")
		})

		Convey("time conversion should be correct", func() {
			c, err := NewCounter("count", 2, nil)
			So(err, ShouldBeNil)

			t := time.Unix(0, 0)
			c.SetTime(t)
			So(*c.pdp.Timestamp, ShouldEqual, 0)
			So(c.Time().Equal(t), ShouldBeTrue)

			t = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
			c.SetTime(t)
			So(*c.pdp.Timestamp, ShouldEqual, 1257894000000)
			So(c.Time().Equal(t), ShouldBeTrue)
		})

		Convey("dimensions should work properly", func() {
			g, err := NewGauge("gauge", 5, map[string]string{
				"a": "1",
				"b": "2",
			})
			So(err, ShouldBeNil)
			So(g, ShouldNotBeNil)
			So(len(g.Dimensions()), ShouldEqual, 2)

			g.SetDimension("", "3")
			So(len(g.Dimensions()), ShouldEqual, 2)

			g.SetDimension("c", "")
			So(len(g.Dimensions()), ShouldEqual, 2)

			g.SetDimension("", "")
			So(len(g.Dimensions()), ShouldEqual, 2)

			g.SetDimension("c", "3")
			So(len(g.Dimensions()), ShouldEqual, 3)

			g.SetDimension("c", "see")
			So(len(g.Dimensions()), ShouldEqual, 3)

			g.SetDimensions(map[string]string{
				"d": "4",
				"e": "5",
			})
			So(len(g.Dimensions()), ShouldEqual, 5)

			g.RemoveDimension("b")
			So(len(g.Dimensions()), ShouldEqual, 4)

			g.RemoveDimension("b")
			So(len(g.Dimensions()), ShouldEqual, 4)

			g.RemoveDimension("d", "a")
			So(len(g.Dimensions()), ShouldEqual, 2)
		})

		Convey("getters should work", func() {
			i := 0
			v := Value(&i)

			c0, err := NewCounter("counter", v, nil)
			So(err, ShouldBeNil)
			So(c0, ShouldNotBeNil)
			So(c0.IntValue(), ShouldEqual, 0)

			i++
			So(c0.IntValue(), ShouldEqual, 1)

			c1, err := NewCounter("counter", 1, nil)
			So(err, ShouldBeNil)
			So(c1, ShouldNotBeNil)
			So(c1.IntValue(), ShouldEqual, 1)

			So(c0.Equal(c1), ShouldBeFalse)

			c2, err := NewCounter("counter", v, nil)
			So(err, ShouldBeNil)
			So(c2, ShouldNotBeNil)
			So(c2.IntValue(), ShouldEqual, 1)

			So(c0.Equal(c2), ShouldBeTrue)
		})

		Convey("WithDimensions should properly clone & add dimensions", func() {
			c1, err := NewCounter("foo", 1, map[string]string{"bar": "quux"})
			So(err, ShouldBeNil)
			c2 := c1.WithDimension("baz", "quuux")
			dims := map[string]string(c1.Dimensions())
			So(dims, ShouldNotContainKey, "foo")
			So(dims, ShouldContainKey, "bar")
			So(dims["bar"], ShouldEqual, "quux")
			So(dims, ShouldNotContainKey, "baz")
			dims = map[string]string(c2.Dimensions())
			So(dims, ShouldNotContainKey, "foo")
			So(dims, ShouldContainKey, "bar")
			So(dims["bar"], ShouldEqual, "quux")
			So(dims, ShouldContainKey, "baz")
			So(dims["baz"], ShouldEqual, "quuux")
		})
	})
}

func ExampleValue() {
	val := 5
	cumulative, _ := NewCumulative("cumulative", Value(&val), nil)

	val = 9
	fmt.Println(cumulative.IntValue())
	// Output: 9
}

func ExampleGetterFunc() {
	val := 5
	count, _ := NewGauge("count", GetterFunc(func() (interface{}, error) {
		return val, nil
	}), nil)

	val++
	fmt.Println(count.IntValue())
	// Output: 6
}
