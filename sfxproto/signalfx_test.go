package sfxproto

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignalFXProto(t *testing.T) {
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
	})
}
