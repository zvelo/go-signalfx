package sfxproto

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/zvelo/go-signalfx/sfxconfig"
)

// DataPoints is a DataPoint set
type DataPoints struct {
	data map[string]*DataPoint
}

// Marshal filters out metrics with empty names, sets a reasonable source on
// each datapoint that doesn't already have a source, filters out dimensions
// with an empty key or value and then marshals the protobuf to a byte slice.
func (dps *DataPoints) Marshal(config *sfxconfig.Config) ([]byte, error) {
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
	*dps = DataPoints{}
}

// Get returns the DataPoint identified by metric
func (dps *DataPoints) Get(metric string) *DataPoint {
	if dp, ok := dps.data[metric]; ok {
		if dp.Metric == metric {
			return dp
		}
	}

	return nil
}

// Del removes the DataPoint identified by metric
func (dps *DataPoints) Del(metric string) {
	delete(dps.data, metric)
}

// Set adds or overwrites a DataPoint identified by metric
func (dps *DataPoints) Set(metric string, value interface{}, metricType MetricType) *DataPoint {
	dp := NewDataPoint(metric, value, metricType)
	dps.data[metric] = dp
	return dp
}
