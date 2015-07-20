package sfxproto

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDataPoints(t *testing.T) {
	p0 := &DataPoint{
		Metric:     proto.String("TestMetric0"),
		Value:      &Datum{IntValue: proto.Int64(99)},
		MetricType: MetricType_COUNTER.Enum(),
	}

	p1 := &DataPoint{
		Metric:     proto.String("TestMetric1"),
		Value:      &Datum{IntValue: proto.Int64(100)},
		MetricType: MetricType_COUNTER.Enum(),
	}

	Convey("Testing DataPoints", t, func() {
		ps := NewDataPoints(2)
		So(ps, ShouldNotBeNil)
		So(ps.Len(), ShouldEqual, 0)
		So(len(ps.data), ShouldEqual, 0)

		So(ps.Add(p0), ShouldEqual, ps)
		So(ps.Add(p1), ShouldEqual, ps)
		So(ps.Len(), ShouldEqual, 2)
		So(len(ps.data), ShouldEqual, 2)

		Convey("List should work", func() {
			list := ps.List()
			So(len(list), ShouldEqual, 2)

		Loop:
			for p0 := range ps.data {
				for _, p1 := range list {
					So(p0, ShouldNotEqual, p1)

					if proto.Equal(p0, p1) {
						continue Loop
					}
				}

				// force fail
				So("didnt find match", ShouldEqual, "")
			}
		})

		Convey("Marshal should work", func() {
			data, err := ps.Marshal()
			So(err, ShouldBeNil)

			msg := &DataPointUploadMessage{}
			err = proto.Unmarshal(data, msg)
			So(err, ShouldBeNil)
			So(len(msg.Datapoints), ShouldEqual, 2)

		Loop:
			for p0 := range ps.data {
				for _, p1 := range msg.Datapoints {
					So(p0, ShouldNotEqual, p1)

					if proto.Equal(p0, p1) {
						continue Loop
					}
				}

				// force fail
				So("didnt find match", ShouldEqual, "")
			}

			tmp := NewDataPoints(0)
			data, err = tmp.Marshal()
			So(data, ShouldBeNil)
			So(err, ShouldEqual, ErrMarshalNoData)
		})

		Convey("Append should work", func() {
			p2 := &DataPoint{
				Metric:     proto.String("TestMetric2"),
				Value:      &Datum{IntValue: proto.Int64(101)},
				MetricType: MetricType_COUNTER.Enum(),
			}

			// make a copy of ps
			ps2 := NewDataPoints(2)
			for p := range ps.data {
				ps2.data[p] = nil
			}

			ps3 := NewDataPoints(1)
			So(ps3, ShouldNotBeNil)
			ps3.Add(p2)
			So(ps3.Len(), ShouldEqual, 1)

			So(ps2.Append(ps3), ShouldEqual, ps2)
			So(ps2.Len(), ShouldEqual, 3)
			So(ps.Len(), ShouldEqual, 2)
		})

		Convey("Remove should work", func() {
			So(ps.Remove(p0), ShouldEqual, ps)
			So(ps.Len(), ShouldEqual, 1)
			for p := range ps.data {
				So(p, ShouldEqual, p1)
			}

			So(ps.Remove(p1, nil), ShouldEqual, ps)
			So(ps.Len(), ShouldEqual, 0)
		})
	})
}
