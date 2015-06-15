package sfxproto

import (
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
)

var (
	// ErrMarshalNoData is returned when marshaling a DataPoints and it has no
	// DataPoint values
	ErrMarshalNoData = fmt.Errorf("no data to marshal")
)

// DataPoints is a DataPoint list
type DataPoints struct {
	data map[*DataPoint]interface{}
	lock sync.Mutex
}

// NewDataPoints creates a new DataPoints object
func NewDataPoints(l int) *DataPoints {
	return &DataPoints{
		data: make(map[*DataPoint]interface{}, l),
	}
}

func (dps *DataPoints) Len() int {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	return len(dps.data)
}

// returns copies for thread safety reasons
func (dps *DataPoints) List() []*DataPoint {
	ret := make([]*DataPoint, 0, dps.Len())

	dps.lock.Lock()
	defer dps.lock.Unlock()

	for dp := range dps.data {
		ret = append(ret, dp.Clone())
	}

	return ret
}

// Marshal filters out metrics with empty names, filters out dimensions with an
// empty or duplicate key or value and then marshals the protobuf to a byte
// slice.
func (dps *DataPoints) Marshal() ([]byte, error) {
	if dps.Len() == 0 {
		return nil, ErrMarshalNoData
	}

	return proto.Marshal(&DataPointUploadMessage{
		Datapoints: dps.List(),
	})
}

// Add a new DataPoint to the list
func (dps *DataPoints) Add(dataPoint *DataPoint) *DataPoints {
	if dataPoint != nil && len(dataPoint.Metric) > 0 {
		dps.lock.Lock()
		defer dps.lock.Unlock()

		dps.data[dataPoint] = nil
	}

	return dps
}

func (dps *DataPoints) Concat(val *DataPoints) *DataPoints {
	val.lock.Lock()
	defer val.lock.Unlock()

	for dp := range val.data {
		dps.Add(dp)
	}

	return dps
}

func (dps *DataPoints) Remove(val *DataPoint) {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	delete(dps.data, val)
}
