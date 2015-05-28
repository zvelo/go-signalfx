package sfxproto

import (
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/zvelo/go-signalfx/sfxconfig"
)

// DataPoints is a DataPoint set
type DataPoints struct {
	data map[string]*DataPoint
	lock sync.RWMutex
}

// Marshal filters out metrics with empty names, sets a reasonable source on
// each datapoint that doesn't already have a source, filters out dimensions
// with an empty key or value and then marshals the protobuf to a byte slice.
func (dps *DataPoints) Marshal(config *sfxconfig.Config) ([]byte, error) {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	filtered := map[string]*DataPoint{}

	for _, dp := range dps.data {
		if len(dp.Metric) == 0 {
			continue
		}

		dp.setReasonableSource(config)
		dp.filterDimensions()

		filtered[dp.Metric] = dp
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

// Reset removes all datapoints from the set
func (dps *DataPoints) Reset() {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	*dps = DataPoints{}
}

// Exists returns if the DataPoint identified by metric is in the set
func (dps *DataPoints) Exists(metric string) bool {
	dps.lock.RLock()
	defer dps.lock.RUnlock()

	if dp, ok := dps.data[metric]; ok {
		if dp.Metric == metric {
			return true
		}
	}

	return false
}

// Del removes the DataPoint identified by metric
func (dps *DataPoints) Del(metric string) {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	delete(dps.data, metric)
}

// Set adds or overwrites a DataPoint identified by metric
func (dps *DataPoints) Set(metricType MetricType, metric string, value interface{}, dimensions Dimensions) {
	dps.lock.Lock()
	defer dps.lock.Unlock()

	dps.data[metric] = NewDataPoint(metricType, metric, value, dimensions)
}
