package sfxproto

import (
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
)

var (
	// ErrMarshalNoData is returned when marshaling a ProtoDataPoints and it has
	// no DataPoint values
	ErrMarshalNoData = fmt.Errorf("no data to marshal")
)

// ProtoDataPoints is a DataPoint list
type ProtoDataPoints struct {
	data map[*ProtoDataPoint]interface{}
	lock sync.Mutex
}

// NewProtoDataPoints creates a new ProtoDataPoints object
func NewProtoDataPoints(l int) *ProtoDataPoints {
	return &ProtoDataPoints{
		data: make(map[*ProtoDataPoint]interface{}, l),
	}
}

func (pdps *ProtoDataPoints) Len() int {
	pdps.lock.Lock()
	defer pdps.lock.Unlock()

	return len(pdps.data)
}

// returns copies for thread safety reasons
func (pdps *ProtoDataPoints) List() []*ProtoDataPoint {
	ret := make([]*ProtoDataPoint, 0, pdps.Len())

	pdps.lock.Lock()
	defer pdps.lock.Unlock()

	for dp := range pdps.data {
		ret = append(ret, dp.Clone())
	}

	return ret
}

// Marshal filters out metrics with empty names, filters out dimensions with an
// empty or duplicate key or value and then marshals the protobuf to a byte
// slice.
func (pdps *ProtoDataPoints) Marshal() ([]byte, error) {
	if pdps.Len() == 0 {
		return nil, ErrMarshalNoData
	}

	return proto.Marshal(&DataPointUploadMessage{
		Datapoints: pdps.List(),
	})
}

// Add a new DataPoint to the list
func (pdps *ProtoDataPoints) Add(dataPoint *ProtoDataPoint) *ProtoDataPoints {
	if dataPoint != nil && dataPoint.Metric != nil && len(*dataPoint.Metric) > 0 {
		pdps.lock.Lock()
		defer pdps.lock.Unlock()

		pdps.data[dataPoint] = nil
	}

	return pdps
}

func (pdps *ProtoDataPoints) Concat(val *ProtoDataPoints) *ProtoDataPoints {
	val.lock.Lock()
	defer val.lock.Unlock()

	for dp := range val.data {
		pdps.Add(dp)
	}

	return pdps
}

func (pdps *ProtoDataPoints) Remove(val *ProtoDataPoint) {
	pdps.lock.Lock()
	defer pdps.lock.Unlock()

	delete(pdps.data, val)
}
