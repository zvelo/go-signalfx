package sfxproto

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
)

func (p *DataPoint) String() string {
	metric := "<nil>"
	if p.Metric != nil {
		metric = *p.Metric
	}

	return fmt.Sprintf("DP[%s\t%s\t%s\t%d\t%s]", metric, p.Dimensions, p.Value, *p.MetricType, p.Time())
}

// Time returns the timestamp of the ProtoDataPoint
func (p *DataPoint) Time() time.Time {
	if p.Timestamp == nil {
		return time.Now()
	}

	return time.Unix(0, *p.Timestamp*int64(time.Millisecond))
}

// SetTime sets the timestamp of the DataPoint
func (p *DataPoint) SetTime(t time.Time) {
	p.Timestamp = proto.Int64(t.UnixNano() / int64(time.Millisecond))
}

// Clone returns a deep copy of the DataPoint
func (p *DataPoint) Clone() *DataPoint {
	return proto.Clone(p).(*DataPoint)
}

// Equal returns whether or not two DataPoint objects are equal
func (p *DataPoint) Equal(val *DataPoint) bool {
	return proto.Equal(p, val)
}
