package signalfx

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type brokenOne struct{}

func (b brokenOne) Get() (interface{}, error) {
	return 0, fmt.Errorf("this is one type of error")
}

func (b brokenOne) Subtract(int64) {
	return
}

type brokenTwo struct{}

func (b brokenTwo) Get() (interface{}, error) {
	return fmt.Errorf("this is a different type of error"), nil
}

func (b brokenTwo) Subtract(int64) {
	return
}

func TestCounter(t *testing.T) {
	Convey("Counters should behave as specified", t, func() {
		c := NewCounter("counter", nil, 0)
		So(c.metric, ShouldEqual, "counter")
		So(c.dimensions, ShouldEqual, nil)
		So(c.value, ShouldEqual, 0)

		So(func() {
			c.PostReportHook(-1)
		}, ShouldPanic)

		cdp := c.DataPoint()
		So(cdp, ShouldBeNil)

		c.Inc(6)
		So(c.value, ShouldEqual, 6)

		cdp = c.DataPoint()
		So(cdp.Metric, ShouldEqual, "counter")
		So(cdp.Dimensions, ShouldEqual, nil)
		So(cdp.Value, ShouldEqual, 6)
		c.PostReportHook(cdp.Value)
		So(c.value, ShouldEqual, 0)
	})

	Convey("Wrapped counters should behave as specified", t, func() {
		c := WrapCounter("counter", nil, brokenOne{})
		So(c, ShouldNotBeNil)
		cdp := c.DataPoint()
		So(cdp, ShouldBeNil)

		c = WrapCounter("counter", nil, brokenTwo{})
		So(c, ShouldNotBeNil)
		cdp = c.DataPoint()
		So(cdp, ShouldBeNil)

		So(func() {
			c.PostReportHook(-12)
		}, ShouldPanic)

		i := NewInt64(0)
		c = WrapCounter("counter", nil, i)
		So(c, ShouldNotBeNil)
		So(*i, ShouldEqual, 0)
		cdp = c.DataPoint()
		So(cdp, ShouldBeNil)

		i.Inc(1)
		cdp = c.DataPoint()
		So(cdp, ShouldNotBeNil)
		So(cdp.Value, ShouldEqual, 1)
		c.PostReportHook(cdp.Value)
		So(*i, ShouldEqual, 0)
	})
}
