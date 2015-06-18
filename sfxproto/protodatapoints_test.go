package sfxproto

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProtoDataPoints(t *testing.T) {
	pdp0 := &ProtoDataPoint{
		Metric:     proto.String("TestMetric0"),
		Value:      &Datum{IntValue: proto.Int64(99)},
		MetricType: MetricType_COUNTER.Enum(),
	}

	pdp1 := &ProtoDataPoint{
		Metric:     proto.String("TestMetric1"),
		Value:      &Datum{IntValue: proto.Int64(100)},
		MetricType: MetricType_COUNTER.Enum(),
	}

	Convey("Testing ProtoDataPoints", t, func() {
		pdps := NewProtoDataPoints(2)
		So(pdps, ShouldNotBeNil)
		So(pdps.Len(), ShouldEqual, 0)
		So(len(pdps.data), ShouldEqual, 0)

		So(pdps.Add(pdp0), ShouldEqual, pdps)
		So(pdps.Add(pdp1), ShouldEqual, pdps)
		So(pdps.Len(), ShouldEqual, 2)
		So(len(pdps.data), ShouldEqual, 2)

		Convey("List should work", func() {
			list := pdps.List()
			So(len(list), ShouldEqual, 2)

		Loop:
			for pdp0 := range pdps.data {
				for _, pdp1 := range list {
					So(pdp0, ShouldNotEqual, pdp1)

					if proto.Equal(pdp0, pdp1) {
						continue Loop
					}
				}

				// force fail
				So("didnt find match", ShouldEqual, "")
			}
		})

		Convey("Marshal should work", func() {
			data, err := pdps.Marshal()
			So(err, ShouldBeNil)

			msg := &DataPointUploadMessage{}
			err = proto.Unmarshal(data, msg)
			So(err, ShouldBeNil)
			So(len(msg.Datapoints), ShouldEqual, 2)

		Loop:
			for pdp0 := range pdps.data {
				for _, pdp1 := range msg.Datapoints {
					So(pdp0, ShouldNotEqual, pdp1)

					if proto.Equal(pdp0, pdp1) {
						continue Loop
					}
				}

				// force fail
				So("didnt find match", ShouldEqual, "")
			}

			tmp := NewProtoDataPoints(0)
			data, err = tmp.Marshal()
			So(data, ShouldBeNil)
			So(err, ShouldEqual, ErrMarshalNoData)
		})

		Convey("Append should work", func() {
			pdp2 := &ProtoDataPoint{
				Metric:     proto.String("TestMetric2"),
				Value:      &Datum{IntValue: proto.Int64(101)},
				MetricType: MetricType_COUNTER.Enum(),
			}

			// make a copy of pdps
			pdps2 := NewProtoDataPoints(2)
			for pdp := range pdps.data {
				pdps2.data[pdp] = nil
			}

			pdps3 := NewProtoDataPoints(1)
			So(pdps3, ShouldNotBeNil)
			pdps3.Add(pdp2)
			So(pdps3.Len(), ShouldEqual, 1)

			So(pdps2.Append(pdps3), ShouldEqual, pdps2)
			So(pdps2.Len(), ShouldEqual, 3)
			So(pdps.Len(), ShouldEqual, 2)
		})

		Convey("Remove should work", func() {
			So(pdps.Remove(pdp0), ShouldEqual, pdps)
			So(pdps.Len(), ShouldEqual, 1)
			for pdp := range pdps.data {
				So(pdp, ShouldEqual, pdp1)
			}

			So(pdps.Remove(pdp1, nil), ShouldEqual, pdps)
			So(pdps.Len(), ShouldEqual, 0)
		})
	})
}
