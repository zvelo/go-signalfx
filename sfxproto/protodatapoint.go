package sfxproto

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
)

func (dp *ProtoDataPoint) String() string {
	metric := "<nil>"
	if dp.Metric != nil {
		metric = *dp.Metric
	}

	return fmt.Sprintf("DP[%s\t%s\t%s\t%d\t%s]", metric, dp.Dimensions, dp.Value, dp.MetricType, dp.Time())
}

func (dp *ProtoDataPoint) Time() time.Time {
	if dp.Timestamp == nil {
		return time.Now()
	}

	return time.Unix(0, *dp.Timestamp*int64(time.Millisecond))
}

func (dp *ProtoDataPoint) SetTime(t time.Time) {
	dp.Timestamp = proto.Int64(t.UnixNano() / int64(time.Millisecond))
}

func (dp *ProtoDataPoint) Clone() *ProtoDataPoint {
	return proto.Clone(dp).(*ProtoDataPoint)
}

func (dp *ProtoDataPoint) Equal(val *ProtoDataPoint) bool {
	return proto.Equal(dp, val)
}
