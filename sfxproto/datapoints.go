package sfxproto

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/zvelo/go-signalfx/sfxconfig"
)

// Marshal filters out metrics with empty names, sets a reasonable source on
// each datapoint that doesn't already have a source, filters out dimensions
// with an empty key or value and then marshals the protobuf to a byte slice.
func (msg *DataPointUploadMessage) Marshal(config *sfxconfig.Config) ([]byte, error) {
	ret := make([]*DataPoint, 0, len(msg.Datapoints))

	for _, dp := range msg.Datapoints {
		if len(dp.Metric) == 0 {
			continue
		}

		dp.setReasonableSource(config)
		dp.filterDimensions()

		ret = append(ret, dp)
	}

	msg.Datapoints = ret

	if len(msg.Datapoints) == 0 {
		return nil, fmt.Errorf("nothing to marshal")
	}

	return proto.Marshal(msg)
}
