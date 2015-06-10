package sfxproto

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func (dp *DataPoint) String() string {
	return fmt.Sprintf("DP[%s\t%s\t%s\t%d\t%s]", dp.Metric, dp.Dimensions, dp.Value, dp.MetricType, dp.Time().String())
}

// Time returns a time.Time object representing the timestamp of the datapoint
func (dp *DataPoint) Time() time.Time {
	return time.Unix(0, dp.Timestamp*int64(time.Millisecond))
}

func (dp *DataPoint) StrValue() string {
	return dp.Value.StrValue
}

func (dp *DataPoint) DoubleValue() float64 {
	return dp.Value.DoubleValue
}

func (dp *DataPoint) IntValue() int64 {
	return dp.Value.IntValue
}

// SetTime properly sets the datapoint timestamp to the millisecond unix epoch
func (dp *DataPoint) SetTime(t time.Time) {
	dp.Timestamp = t.UnixNano() / int64(time.Millisecond)
}

// NewDataPoint creates a new datapoint
func NewDataPoint(metricType MetricType, metric string, value interface{}, dimensions Dimensions) (*DataPoint, error) {
	ret := &DataPoint{
		Metric:     metric,
		MetricType: metricType,
	}

	if dimensions != nil {
		ret.Dimensions = dimensions.List()
	}

	ret.SetTime(time.Now())

	if err := ret.Set(value); err != nil {
		return nil, err
	}

	return ret, nil
}

func NewCumulative(metricName string, value interface{}, defaultDims Dimensions) (*DataPoint, error) {
	return NewDataPoint(MetricType_CUMULATIVE_COUNTER, metricName, value, defaultDims)
}

func NewGauge(metricName string, value interface{}, defaultDims Dimensions) (*DataPoint, error) {
	return NewDataPoint(MetricType_GAUGE, metricName, value, defaultDims)
}

func NewCounter(metricName string, value interface{}, defaultDims Dimensions) (*DataPoint, error) {
	return NewDataPoint(MetricType_COUNTER, metricName, value, defaultDims)
}

// Set sets the datapoint value datum correctly for all integer, float and
// string types. If another type is passed in, an error is returned.
func (dp *DataPoint) Set(val interface{}) error {
	dp.Value = &Datum{}
	return dp.Value.Set(val)
}

func massageKey(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' {
			return r
		}

		return '_'
	}, str)
}
