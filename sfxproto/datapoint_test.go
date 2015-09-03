package sfxproto

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDataPoint(t *testing.T) {
	p := &DataPoint{
		Metric: proto.String("TestMetric"),
		Value: &Datum{
			IntValue: proto.Int64(99),
		},
		MetricType: MetricType_COUNTER.Enum(),
		Dimensions: []*Dimension{
			{Key: proto.String("dim0"), Value: proto.String("val0")},
			{Key: proto.String("dim1"), Value: proto.String("val1")},
		},
	}

	Convey("Testing ProtoDataPoint", t, func() {
		now := time.Now()

		So(now.Before(p.Time()), ShouldBeTrue)

		p.SetTime(now)

		// intentially switch to ms resolution
		now = time.Unix(0, now.UnixNano()/int64(time.Millisecond)*int64(time.Millisecond))

		So(p.Time().Equal(now), ShouldBeTrue)

		So(p.String(), ShouldEqual, "DP[TestMetric\t[key:\"dim0\" value:\"val0\"  key:\"dim1\" value:\"val1\" ]\tintValue:99 \t1\t"+now.String()+"]")

		clone := p.Clone()
		So(p, ShouldNotEqual, clone)
		So(proto.Equal(p, clone), ShouldBeTrue)
		So(p.Equal(clone), ShouldBeTrue)
		So(clone.Equal(p), ShouldBeTrue)
	})
}
