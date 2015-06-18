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

// ProtoDataPoints is a set of ProtoDataPoint objects
type ProtoDataPoints struct {
	data map[*ProtoDataPoint]interface{}
	lock sync.Mutex
}

// NewProtoDataPoints creates a new ProtoDataPoints object with expected size of
// l
func NewProtoDataPoints(l int) *ProtoDataPoints {
	return &ProtoDataPoints{
		data: make(map[*ProtoDataPoint]interface{}, l),
	}
}

// Len returns the number of ProtoDataPoint objects the ProtoDataPoints object contains
func (pdps *ProtoDataPoints) Len() int {
	pdps.lock.Lock()
	defer pdps.lock.Unlock()

	return len(pdps.data)
}

// List returns a slice of copies of each ProtoDataPoint
func (pdps *ProtoDataPoints) List() []*ProtoDataPoint {
	ret := make([]*ProtoDataPoint, 0, pdps.Len())

	pdps.lock.Lock()
	defer pdps.lock.Unlock()

	for pdp := range pdps.data {
		ret = append(ret, pdp.Clone())
	}

	return ret
}

// Marshal a DataPointUploadMessage from the ProtoDataPoints object
func (pdps *ProtoDataPoints) Marshal() ([]byte, error) {
	if pdps.Len() == 0 {
		return nil, ErrMarshalNoData
	}

	return proto.Marshal(&DataPointUploadMessage{
		Datapoints: pdps.List(),
	})
}

// Add a new DataPoint to the list
func (pdps *ProtoDataPoints) Add(protoDataPoint *ProtoDataPoint) *ProtoDataPoints {
	if protoDataPoint != nil && protoDataPoint.Metric != nil && len(*protoDataPoint.Metric) > 0 {
		pdps.lock.Lock()
		defer pdps.lock.Unlock()

		pdps.data[protoDataPoint] = nil
	}

	return pdps
}

// Append appends the passed ProtoDataPoints to the source object
func (pdps *ProtoDataPoints) Append(val *ProtoDataPoints) *ProtoDataPoints {
	val.lock.Lock()
	defer val.lock.Unlock()

	for pdp := range val.data {
		pdps.Add(pdp)
	}

	return pdps
}

// Remove ProtoDataPoint(s) from the set
func (pdps *ProtoDataPoints) Remove(vals ...*ProtoDataPoint) {
	pdps.lock.Lock()
	defer pdps.lock.Unlock()

	for _, val := range vals {
		delete(pdps.data, val)
	}
}
