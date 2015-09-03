package sfxproto

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/ztag/vendor/github.com/gogo/protobuf/proto"
)

func TestSignalFXProto(t *testing.T) {
	p := &DataPoint{
		Metric:     proto.String("metric0"),
		MetricType: MetricType_CUMULATIVE_COUNTER.Enum(),
		Value:      &Datum{IntValue: proto.Int64(64)},
		Dimensions: []*Dimension{
			{Key: proto.String("key0"), Value: proto.String("value0")},
		},
	}
	now := time.Now()
	p.SetTime(now)
	now = time.Unix(0, now.UnixNano()/int64(time.Millisecond))

	Convey("Testing SignalFX Proto", t, func() {
		Convey("MetricType", func() {
			mt := MetricType_GAUGE
			So(mt.String(), ShouldEqual, "GAUGE")

			data, err := proto.MarshalJSONEnum(MetricType_name, int32(mt))
			So(err, ShouldBeNil)

			var mt2 MetricType
			err = mt2.UnmarshalJSON(data)
			So(err, ShouldBeNil)
			So(mt, ShouldEqual, mt2)

			err = mt2.UnmarshalJSON([]byte("{}"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "cannot unmarshal `{}` into enum MetricType")
		})

		Convey("Datum", func() {
			d := &Datum{
				IntValue: proto.Int64(16),
			}

			So(d.StrValue, ShouldBeNil)
			So(d.DoubleValue, ShouldBeNil)
			So(*d.IntValue, ShouldEqual, 16)

			d.Reset()
			So(d.StrValue, ShouldBeNil)
			So(d.DoubleValue, ShouldBeNil)
			So(d.IntValue, ShouldBeNil)

			d.ProtoMessage() // noop

			So(d.GetStrValue(), ShouldEqual, "")
			So(d.GetDoubleValue(), ShouldEqual, 0)
			So(d.GetIntValue(), ShouldEqual, 0)

			d.IntValue = proto.Int64(8)
			So(d.GetIntValue(), ShouldEqual, 8)

			d.StrValue = proto.String("something")
			So(d.GetStrValue(), ShouldEqual, "something")

			d.DoubleValue = proto.Float64(32)
			So(d.GetDoubleValue(), ShouldEqual, 32)
		})

		Convey("Dimension", func() {
			d := &Dimension{Key: proto.String("key0"), Value: proto.String("value0")}
			So(d.GetKey(), ShouldEqual, "key0")
			So(d.GetValue(), ShouldEqual, "value0")
			d.ProtoMessage() // noop
			So(d.String(), ShouldEqual, `key:"key0" value:"value0" `)
			d.Reset()
			So(d.GetKey(), ShouldEqual, "")
			So(d.GetValue(), ShouldEqual, "")
		})

		Convey("DataPointUploadMessage", func() {
			p := p.Clone()
			dps := []*DataPoint{p}
			dpum := &DataPointUploadMessage{
				Datapoints: dps,
			}
			ts := fmt.Sprintf("%d", now.UnixNano())
			So(dpum.String(), ShouldEqual, `datapoints:<metric:"metric0" timestamp:`+ts+` value:<intValue:64 > metricType:CUMULATIVE_COUNTER dimensions:<key:"key0" value:"value0" > > `)
			dpum.ProtoMessage() // noop
			So(dpum.GetDatapoints(), ShouldResemble, dps)
			dpum.Reset()
			So(dpum.GetDatapoints(), ShouldResemble, []*DataPoint(nil))
			dpum = nil
			So(dpum.GetDatapoints(), ShouldBeNil)
		})

		Convey("DataPoint", func() {
			p := p.Clone()
			p.ProtoMessage() // noop

			So(p.GetMetric(), ShouldEqual, "metric0")
			So(p.GetTimestamp(), ShouldEqual, now.UnixNano())
			So(*p.GetValue().IntValue, ShouldEqual, 64)
			So(p.GetMetricType(), ShouldEqual, MetricType_CUMULATIVE_COUNTER)
			So(p.GetDimensions()[0].GetKey(), ShouldEqual, "key0")

			p.Reset()

			So(p.GetMetric(), ShouldEqual, "")
			So(p.GetTimestamp(), ShouldEqual, 0)
			So(p.GetValue(), ShouldBeNil)
			So(p.GetMetricType(), ShouldEqual, MetricType_GAUGE)
			So(p.GetDimensions(), ShouldBeNil)

			p = nil
			So(p.GetValue(), ShouldBeNil)
			So(p.GetDimensions(), ShouldBeNil)
		})
	})
}
