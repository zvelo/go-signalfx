package sfxproto

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	p2 "github.com/signalfx/com_signalfx_metrics_protobuf"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDataPoint(t *testing.T) {
	Convey("Testing DataPoint", t, func() {
		Convey("ensuring proto3 version matches proto2", func() {
			p2EQp3 := func(p2dp *p2.DataPoint, p3dp *DataPoint) {
				So(p3dp.Metric, ShouldEqual, *p2dp.Metric)
				So(p3dp.MetricType, ShouldEqual, *p2dp.MetricType)
				So(p3dp.Value.StrValue, ShouldEqual, *p2dp.Value.StrValue)
				So(p3dp.Value.DoubleValue, ShouldEqual, *p2dp.Value.DoubleValue)
				So(p3dp.Value.IntValue, ShouldEqual, *p2dp.Value.IntValue)
				So(p3dp.Timestamp, ShouldEqual, *p2dp.Timestamp)
				So(len(p3dp.Dimensions), ShouldEqual, len(p2dp.Dimensions))

				for _, p3dim := range p3dp.Dimensions {
					for _, p2dim := range p2dp.Dimensions {
						if p3dim.Key == *p2dim.Key {
							So(p3dim.Value, ShouldEqual, *p2dim.Value)
						}
					}
				}
			}

			t := time.Now()
			dp := &DataPoint{
				Metric:     "test_data",
				MetricType: MetricType_COUNTER,
				Value: &Datum{
					IntValue: 123,
				},
				Dimensions: []*Dimension{
					{Key: "key0", Value: "value0"},
					{Key: "key1", Value: "value1"},
				},
			}
			dp.SetTime(t)

			Convey("MetricType values", func() {
				So(MetricType_GAUGE, ShouldEqual, p2.MetricType_GAUGE)
				So(MetricType_COUNTER, ShouldEqual, p2.MetricType_COUNTER)
				So(MetricType_ENUM, ShouldEqual, p2.MetricType_ENUM)
				So(MetricType_CUMULATIVE_COUNTER, ShouldEqual, p2.MetricType_CUMULATIVE_COUNTER)
				So(len(MetricType_name), ShouldEqual, len(p2.MetricType_name))
			})

			Convey("marshaling p3 to p2 works", func() {
				p3dp := dp.Clone()

				// marshal a proto3 datapoint to bytes
				data, err := proto.Marshal(p3dp)
				So(err, ShouldBeNil)
				So(data, ShouldNotBeNil)

				// unmarshal the bytes into a proto2 datapoint
				p2dp := &p2.DataPoint{}
				err = proto.Unmarshal(data, p2dp)
				So(err, ShouldBeNil)

				// ensure they are equal
				p2EQp3(p2dp, p3dp)
			})

			Convey("marshaling p2 to p3 works", func() {
				p2dp := &p2.DataPoint{
					Metric:     proto.String("test_data"),
					MetricType: p2.MetricType_COUNTER.Enum(),
					Value: &p2.Datum{
						IntValue: proto.Int64(123),
					},
					Timestamp: proto.Int64(dp.Timestamp),
					Dimensions: []*p2.Dimension{
						{Key: proto.String("key0"), Value: proto.String("value0")},
						{Key: proto.String("key1"), Value: proto.String("value1")},
					},
				}
				So(p2dp, ShouldNotBeNil)

				// marshal a proto2 datapoint to bytes
				data, err := proto.Marshal(p2dp)
				So(err, ShouldBeNil)
				So(data, ShouldNotBeNil)

				// unmarshal the bytes into a proto2 datapoint
				p3dp := &DataPoint{}
				err = proto.Unmarshal(data, p3dp)
				So(err, ShouldBeNil)

				// ensure they are equal
				So(dp.Equal(p3dp), ShouldBeTrue)
			})
		})
	})
}
