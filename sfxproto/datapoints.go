package sfxproto

import (
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/zvelo/go-signalfx/sfxconfig"
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
	return len(dps.data)
}

// Marshal filters out metrics with empty names, filters out dimensions with an
// empty or duplicate key or value and then marshals the protobuf to a byte
// slice.
func (dps *DataPoints) Marshal(config *sfxconfig.Config) ([]byte, error) {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	if dps.Len() == 0 {
		return nil, ErrMarshalNoData
	}

	ret := DataPointUploadMessage{}
	for dp := range dps.data {
		ret.Datapoints = append(ret.Datapoints, dp)
	}

	return proto.Marshal(&ret)
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

func (dps *DataPoints) NewCumulative(metricName string, value interface{}, dims Dimensions) (*DataPoint, error) {
	dp, err := NewCumulative(metricName, value, dims)
	if err != nil {
		return nil, err
	}

	dps.Add(dp)
	return dp, nil
}

func (dps *DataPoints) NewGauge(metricName string, value interface{}, dims Dimensions) (*DataPoint, error) {
	dp, err := NewGauge(metricName, value, dims)
	if err != nil {
		return nil, err
	}

	dps.Add(dp)
	return dp, err
}

func (dps *DataPoints) NewCounter(metricName string, value interface{}, dims Dimensions) (*DataPoint, error) {
	dp, err := NewCounter(metricName, value, dims)
	if err != nil {
		return nil, err
	}

	dps.Add(dp)
	return dp, err
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
