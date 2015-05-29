package sfxproto

import (
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/zvelo/go-signalfx/sfxconfig"
)

// DataPoints is a DataPoint list
type DataPoints struct {
	data []*DataPoint
	lock sync.Mutex
}

// NewDataPoints creates a new DataPoints object
func NewDataPoints() *DataPoints {
	return &DataPoints{}
}

// Marshal filters out metrics with empty names, sets a reasonable source on
// each datapoint that doesn't already have a source, filters out dimensions
// with an empty key or value and then marshals the protobuf to a byte slice.
func (dps *DataPoints) Marshal(config *sfxconfig.Config) ([]byte, error) {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	filtered := make([]*DataPoint, len(dps.data))

	for _, dp := range dps.data {
		if len(dp.Metric) == 0 {
			continue
		}

		dp.setReasonableSource(config)
		dp.filterDimensions()

		filtered = append(filtered, dp)
	}

	dps.data = filtered

	if len(filtered) == 0 {
		return nil, fmt.Errorf("nothing to marshal")
	}

	ret := DataPointUploadMessage{}
	for _, dp := range filtered {
		ret.Datapoints = append(ret.Datapoints, dp)
	}

	return proto.Marshal(&ret)
}

// Add a new DataPoint to the list
func (dps *DataPoints) Add(dataPoint *DataPoint) *DataPoints {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	dps.data = append(dps.data, dataPoint)
	return dps
}
