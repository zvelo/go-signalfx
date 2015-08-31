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
			c := NewCounter("count", nil, 2)
			cdp := c.DataPoint()
			So(cdp.Value, ShouldEqual, 2)
			So(cdp.Metric, ShouldEqual, "count")
			So(cdp.Timestamp.Before(time.Now()), ShouldBeTrue)
			So(cdp.Type, ShouldEqual, sfxproto.MetricType_COUNTER)

			dim := c.dimensions
			So(dim, ShouldBeNil)
			So(len(dim), ShouldEqual, 0)
		})

		Convey("created datapoint should have the correct value", func() {
			m0 := NewCounter("count", nil, 3)
			So(m0, ShouldNotBeNil)
			So(m0.value, ShouldEqual, 3)
		})

		Convey("time conversion should be correct", func() {
			c := NewCounter("count", nil, 2)
			So(c, ShouldNotBeNil)

			cdp := c.DataPoint()
			So(cdp, ShouldNotBeNil)
			t := time.Unix(0, 0)
			cdp.Timestamp = t
			cpdp := cdp.protoDataPoint("", nil)
			So(cpdp, ShouldNotBeNil)
			So(*cpdp.Timestamp, ShouldEqual, 0)

			t = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
			cdp.Timestamp = t
			cpdp = cdp.protoDataPoint("", nil)
			So(*cpdp.Timestamp, ShouldEqual, 1257894000000)
		})

		Convey("dimensions should work properly", func() {
			g := NewGauge(
				"gauge",
				map[string]string{
					"a": "1",
					"b": "2",
				},
				5,
			)
			So(g, ShouldNotBeNil)
			So(len(g.dimensions), ShouldEqual, 2)
		})

		Convey("getters should work", func() {
			i := NewInt64(0)
			v := Subtractor(i)

			c0 := WrapCounter("counter", nil, v)
			So(c0, ShouldNotBeNil)
			cdp0 := c0.DataPoint()
			So(cdp0.Value, ShouldEqual, 0)

			*i++
			cdp0 = c0.DataPoint()
			So(cdp0.Value, ShouldEqual, 1)

			c1 := WrapCounter("counter", nil, NewInt64(1))
			So(c1, ShouldNotBeNil)
			cdp1 := c1.DataPoint()
			So(cdp1.Value, ShouldEqual, 1)

			So(cdp0.Value == cdp1.Value, ShouldBeTrue)

			c2 := WrapCounter("counter", nil, v)
			So(c2, ShouldNotBeNil)
			cdp2 := c2.DataPoint()
			So(cdp2.Value, ShouldEqual, 1)

			So(cdp0.Value == cdp2.Value, ShouldBeTrue)
		})
	})
}

func ExampleValue() {
	val := 5
	cumulative := WrapCumulativeCounter("cumulative", nil, Value(&val))

	val = 9
	cdp := cumulative.DataPoint()
	fmt.Println(cdp.Value)
	// Output: 9
}

func ExampleGetterFunc() {
	val := 5
	count := WrapGauge("count", nil, GetterFunc(func() (interface{}, error) {
		return val, nil
	}))

	val++
	cdp := count.DataPoint()
	fmt.Println(cdp.Value)
	// Output: 6
}
