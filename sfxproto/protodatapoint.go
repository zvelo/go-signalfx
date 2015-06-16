package sfxproto

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
)

func (pdp *ProtoDataPoint) String() string {
	metric := "<nil>"
	if pdp.Metric != nil {
		metric = *pdp.Metric
	}

	return fmt.Sprintf("DP[%s\t%s\t%s\t%d\t%s]", metric, pdp.Dimensions, pdp.Value, pdp.MetricType, pdp.Time())
}

// Time returns the timestamp of the ProtoDataPoint
func (pdp *ProtoDataPoint) Time() time.Time {
	if pdp.Timestamp == nil {
		return time.Now()
	}

	return time.Unix(0, *pdp.Timestamp*int64(time.Millisecond))
}

// SetTime sets the timestamp of the ProtoDataPoint
func (pdp *ProtoDataPoint) SetTime(t time.Time) {
	pdp.Timestamp = proto.Int64(t.UnixNano() / int64(time.Millisecond))
}

// Clone returns a deep copy of the ProtoDataPoint
func (pdp *ProtoDataPoint) Clone() *ProtoDataPoint {
	return proto.Clone(pdp).(*ProtoDataPoint)
}

// Equal returns whether or not two ProtoDataPoint objects are equal
func (pdp *ProtoDataPoint) Equal(val *ProtoDataPoint) bool {
	return proto.Equal(pdp, val)
}
