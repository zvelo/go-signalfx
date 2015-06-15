package sfxproto

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
)

func (dp *DataPoint) String() string {
	return fmt.Sprintf("DP[%s\t%s\t%s\t%d\t%s]", dp.Metric, dp.Dimensions, dp.Value, dp.MetricType, dp.Time())
}

func (dp *DataPoint) Time() time.Time {
	return time.Unix(0, dp.Timestamp*int64(time.Millisecond))
}

func (dp *DataPoint) SetTime(t time.Time) {
	dp.Timestamp = t.UnixNano() / int64(time.Millisecond)
}

func (dp *DataPoint) Clone() *DataPoint {
	return proto.Clone(dp).(*DataPoint)
}

func (dp *DataPoint) Equal(val *DataPoint) bool {
	return proto.Equal(dp, val)
}
