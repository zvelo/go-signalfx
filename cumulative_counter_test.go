package signalfx

import (
	"fmt"
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCumulativeCounter(t *testing.T) {
	Convey("Cumulative counters should behave as specified", t, func() {
		cc := NewCumulativeCounter("cumulative-counter", nil, 0)
		So(cc, ShouldNotBeNil)
		So(cc.metric, ShouldEqual, "cumulative-counter")
		So(cc.dimensions, ShouldEqual, nil)
		So(cc.value, ShouldEqual, 0)

		Convey("Cumulative counters are bounded by MaxInt64", func() {
			cc.Sample(math.MaxInt64 + 1)
			ccdp := cc.DataPoint()
			So(ccdp, ShouldBeNil)
		})

		Convey("Cumulative counters must be non-negative", func() {
			So(func() {
				cc.PostReportHook(-12)
			}, ShouldPanic)
		})

		cc.Sample(16)
		So(cc.value, ShouldEqual, 16)
		So(cc.previousValue, ShouldEqual, 0)
		ccdp := cc.DataPoint()
		So(cc.previousValue, ShouldEqual, 0)
		So(ccdp, ShouldNotBeNil)
		So(ccdp.Metric, ShouldEqual, "cumulative-counter")
		So(ccdp.Dimensions, ShouldEqual, nil)
		So(ccdp.Value, ShouldEqual, 16)
		cc.PostReportHook(ccdp.Value)
		So(cc.previousValue, ShouldEqual, 16)
		ccdp = cc.DataPoint()
		So(ccdp, ShouldBeNil)
		cc.Sample(24)
		cc.PostReportHook(18)
		So(cc.previousValue, ShouldEqual, 18)
		cc.PostReportHook(16)
		So(cc.previousValue, ShouldEqual, 18)
		ccdp = cc.DataPoint()
		So(ccdp, ShouldNotBeNil)
		So(ccdp.Value, ShouldEqual, 24)
	})
	Convey("Broken wrapped cumulative counters break cleanly", t, func() {
		cc := WrapCumulativeCounter("broken", nil, GetterFunc(func() (interface{}, error) {
			return 0, fmt.Errorf("this is one type of error")
		}))
		So(cc, ShouldNotBeNil)
		ccdp := cc.DataPoint()
		So(ccdp, ShouldBeNil)

		cc = WrapCumulativeCounter("broken", nil, GetterFunc(func() (interface{}, error) {
			return fmt.Errorf("this is a different type of error"), nil
		}))
		So(cc, ShouldNotBeNil)
		ccdp = cc.DataPoint()
		So(ccdp, ShouldBeNil)

		cc = WrapCumulativeCounter("broken", nil, GetterFunc(func() (interface{}, error) {
			return -1, nil
		}))
		So(cc, ShouldNotBeNil)
		ccdp = cc.DataPoint()
		So(ccdp, ShouldBeNil)

		So(func() {
			cc.PostReportHook(-1)
		}, ShouldPanic)
		cc.previousValue = 16
		cc.PostReportHook(15)
		So(cc.previousValue, ShouldEqual, 16)
		cc.PostReportHook(17)
		So(cc.previousValue, ShouldEqual, 17)
	})
}
