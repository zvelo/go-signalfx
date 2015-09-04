package signalfx

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGauge(t *testing.T) {
	Convey("Gauge works as specified", t, func() {
		g := NewGauge("gauge", nil, 0)
		So(g, ShouldNotBeNil)
		So(g.metric, ShouldEqual, "gauge")
		So(g.dimensions, ShouldBeNil)
		So(g.value, ShouldEqual, 0)

		g.Record(12)
		So(g.value, ShouldEqual, 12)

		gdp := g.DataPoint()
		So(gdp, ShouldNotBeNil)
		So(gdp.Metric, ShouldEqual, "gauge")
		So(gdp.Dimensions, ShouldBeNil)
		So(gdp.Value, ShouldEqual, 12)
		t := gdp.Timestamp

		// calling g.DataPoint repeatedly should yield similar
		// data points
		time.Sleep(time.Millisecond)
		gdp = g.DataPoint()
		So(gdp, ShouldNotBeNil)
		So(gdp.Metric, ShouldEqual, "gauge")
		So(gdp.Dimensions, ShouldBeNil)
		So(gdp.Value, ShouldEqual, 12)
		So(t.Before(gdp.Timestamp), ShouldBeTrue)
	})
	Convey("Broken wrapped gauges break cleanly", t, func() {
		g := WrapGauge("broken", nil, GetterFunc(func() (interface{}, error) {
			return 0, fmt.Errorf("this is an error")
		}))
		gdp := g.DataPoint()
		So(gdp, ShouldBeNil)
		g = WrapGauge("broken", nil, GetterFunc(func() (interface{}, error) {
			return fmt.Errorf("this is a different type of error"), nil
		}))
		gdp = g.DataPoint()
		So(gdp, ShouldBeNil)
	})
}
