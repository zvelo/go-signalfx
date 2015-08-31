package signalfx

import (
	"fmt"
	"time"

	"github.com/zvelo/go-signalfx/sfxproto"
)

var (
	// ErrIllegalType is returned when trying to set the Datum value using an
	// unsupported type
	ErrIllegalType = fmt.Errorf("illegal value type")

	// ErrNoMetricName is returned when trying to create a DataPoint without a
	// Metric name
	ErrNoMetricName = fmt.Errorf("no metric name")
)

type MetricType sfxproto.MetricType

const (
	CounterType           MetricType = MetricType(sfxproto.MetricType_COUNTER)
	CumulativeCounterType MetricType = MetricType(sfxproto.MetricType_CUMULATIVE_COUNTER)
	GaugeType             MetricType = MetricType(sfxproto.MetricType_GAUGE)
)

type DataPoint struct {
	Metric     string
	Type       MetricType
	Value      int64
	Timestamp  time.Time
	Dimensions map[string]string
}

func (dp DataPoint) protoDataPoint(
	metricPrefix string,
	dimensions []*sfxproto.Dimension,
) *sfxproto.DataPoint {
	timestamp := dp.Timestamp.UnixNano() / 1000000
	fullDims := make(
		[]*sfxproto.Dimension,
		len(dimensions),
		len(dimensions)+len(dp.Dimensions))
	for i, v := range dimensions {
		fullDims[i] = v
	}
	for k, v := range dp.Dimensions {
		// have to copy the values, since these are stored as
		// pointers…
		var dk, dv string
		dk = k
		dv = v
		fullDims = append(
			fullDims,
			&sfxproto.Dimension{Key: &dk, Value: &dv},
		)
	}
	metric := metricPrefix + dp.Metric
	return &sfxproto.DataPoint{
		Metric:     &metric,
		Timestamp:  &timestamp,
		Value:      &sfxproto.Datum{IntValue: &dp.Value},
		Dimensions: fullDims,
	}
}
