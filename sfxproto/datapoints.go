package sfxproto

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
)

var (
	// ErrMarshalNoData is returned when marshaling a DataPoints and it has
	// no DataPoint values
	ErrMarshalNoData = fmt.Errorf("no data to marshal")
)

// DataPoints is a set of DataPoint objects
type DataPoints struct {
	data map[*DataPoint]interface{}
	lock sync.Mutex
}

// NewDataPoints creates a new DataPoints object with expected size of
// l
func NewDataPoints(l int) *DataPoints {
	return &DataPoints{
		data: make(map[*DataPoint]interface{}, l),
	}
}

// Len returns the number of DataPoint objects the DataPoints object contains
func (ps *DataPoints) Len() int {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	return len(ps.data)
}

// List returns a slice of copies of each DataPoint
func (ps *DataPoints) List() []*DataPoint {
	ret := make([]*DataPoint, 0, ps.Len())

	ps.lock.Lock()
	defer ps.lock.Unlock()

	for p := range ps.data {
		ret = append(ret, p.Clone())
	}

	return ret
}

// Marshal a DataPointUploadMessage from the DataPoints object
func (ps *DataPoints) Marshal() ([]byte, error) {
	if ps == nil || ps.Len() == 0 {
		return nil, ErrMarshalNoData
	}

	return proto.Marshal(&DataPointUploadMessage{
		Datapoints: ps.List(),
	})
}

// Add a new DataPoint to the list
func (ps *DataPoints) Add(dataPoint *DataPoint) *DataPoints {
	if dataPoint != nil && dataPoint.Metric != nil && len(*dataPoint.Metric) > 0 {
		ps.lock.Lock()
		defer ps.lock.Unlock()

		ps.data[dataPoint] = nil
	}

	return ps
}

// Append appends the passed DataPoints to the source object
func (ps *DataPoints) Append(val *DataPoints) *DataPoints {
	val.lock.Lock()
	defer val.lock.Unlock()

	for p := range val.data {
		ps.Add(p)
	}

	return ps
}

// Remove DataPoint(s) from the set
func (ps *DataPoints) Remove(vals ...*DataPoint) *DataPoints {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, val := range vals {
		delete(ps.data, val)
	}

	return ps
}
