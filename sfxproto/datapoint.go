package sfxproto

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/zvelo/go-signalfx/sfxconfig"
)

func (dp *DataPoint) String() string {
	return fmt.Sprintf("DP[%s\t%s\t%s\t%d\t%s]", dp.Metric, dp.Dimensions, dp.Value, dp.MetricType, dp.Time().String())
}

// Time returns a time.Time object representing the timestamp of the datapoint
func (dp *DataPoint) Time() time.Time {
	return time.Unix(0, dp.Timestamp*int64(time.Millisecond))
}

// SetTime properly sets the datapoint timestamp to the millisecond unix epoch
func (dp *DataPoint) SetTime(t time.Time) {
	dp.Timestamp = t.UnixNano() / int64(time.Millisecond)
}

func (dp *DataPoint) setReasonableSource(config *sfxconfig.Config) {
	if len(dp.Source) > 0 {
		return
	}

	for _, sourceName := range config.DimensionSources {
		for _, dimension := range dp.Dimensions {
			if sourceName == dimension.Key && len(dimension.Value) > 0 {
				dp.Source = dimension.Value
				return
			}
		}
	}

	// TODO(jrubin) what if this is empty?
	dp.Source = config.DefaultSource
}

// NewDataPoint creates a new datapoint
func NewDataPoint(metricType MetricType, metric string, value interface{}, timeStamp time.Time, dimensions Dimensions) *DataPoint {
	// TODO(jrubin) what about Source?

	ret := &DataPoint{
		Metric:     metric,
		MetricType: metricType,
		Dimensions: dimensions.Clone(),
	}

	ret.SetTime(timeStamp)

	if err := ret.SetValue(value); err != nil {
		return nil
	}

	return ret
}

// SetValue sets the datapoint value datum correctly for all integer, float and
// string types. If another type is passed in, an error is returned.
func (dp *DataPoint) SetValue(val interface{}) error {
	dp.Value = &Datum{}
	return dp.Value.Set(val)
}

func (dp *DataPoint) filterDimensions() {
	ret := make([]*Dimension, 0, len(dp.Dimensions))
	for _, dimension := range dp.Dimensions {
		if dimension.Key == "" || dimension.Value == "" {
			continue
		}

		dimension.Key = massageKey(dimension.Key)
		ret = append(ret, dimension)
	}

	dp.Dimensions = ret
}

func massageKey(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' {
			return r
		}

		return '_'
	}, str)
}
