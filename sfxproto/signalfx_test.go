package sfxproto

import (
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignalFXProto(t *testing.T) {
	pdp := &ProtoDataPoint{
		Metric:     proto.String("metric0"),
		MetricType: MetricType_CUMULATIVE_COUNTER.Enum(),
		Value:      &Datum{IntValue: proto.Int64(64)},
		Dimensions: []*Dimension{
			{Key: proto.String("key0"), Value: proto.String("value0")},
		},
	}
	now := time.Now()
	pdp.SetTime(now)
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
			pdp := pdp.Clone()
			dps := []*ProtoDataPoint{pdp}
			dpum := &DataPointUploadMessage{
				Datapoints: dps,
			}
			ts := fmt.Sprintf("%d", now.UnixNano())
			So(dpum.String(), ShouldEqual, `datapoints:<metric:"metric0" timestamp:`+ts+` value:<intValue:64 > metricType:CUMULATIVE_COUNTER dimensions:<key:"key0" value:"value0" > > `)
			dpum.ProtoMessage() // noop
			So(dpum.GetDatapoints(), ShouldResemble, dps)
			dpum.Reset()
			So(dpum.GetDatapoints(), ShouldResemble, []*ProtoDataPoint(nil))
			dpum = nil
			So(dpum.GetDatapoints(), ShouldBeNil)
		})

		Convey("ProtoDataPoint", func() {
			pdp := pdp.Clone()
			pdp.ProtoMessage() // noop

			So(pdp.GetMetric(), ShouldEqual, "metric0")
			So(pdp.GetTimestamp(), ShouldEqual, now.UnixNano())
			So(*pdp.GetValue().IntValue, ShouldEqual, 64)
			So(pdp.GetMetricType(), ShouldEqual, MetricType_CUMULATIVE_COUNTER)
			So(pdp.GetDimensions()[0].GetKey(), ShouldEqual, "key0")

			pdp.Reset()

			So(pdp.GetMetric(), ShouldEqual, "")
			So(pdp.GetTimestamp(), ShouldEqual, 0)
			So(pdp.GetValue(), ShouldBeNil)
			So(pdp.GetMetricType(), ShouldEqual, MetricType_GAUGE)
			So(pdp.GetDimensions(), ShouldBeNil)

			pdp = nil
			So(pdp.GetValue(), ShouldBeNil)
			So(pdp.GetDimensions(), ShouldBeNil)
		})
	})
}
