package sfxproto

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProtoDataPoint(t *testing.T) {
	pdp := &ProtoDataPoint{
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

		So(now.Before(pdp.Time()), ShouldBeTrue)

		pdp.SetTime(now)

		// intentially switch to ms resolution
		now = time.Unix(0, now.UnixNano()/int64(time.Millisecond)*int64(time.Millisecond))

		So(pdp.Time().Equal(now), ShouldBeTrue)

		So(pdp.String(), ShouldEqual, "DP[TestMetric\t[key:\"dim0\" value:\"val0\"  key:\"dim1\" value:\"val1\" ]\tintValue:99 \t1\t"+now.String()+"]")

		clone := pdp.Clone()
		So(pdp, ShouldNotEqual, clone)
		So(proto.Equal(pdp, clone), ShouldBeTrue)
		So(pdp.Equal(clone), ShouldBeTrue)
		So(clone.Equal(pdp), ShouldBeTrue)
	})
}
