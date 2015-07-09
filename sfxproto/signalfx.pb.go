// Code generated by protoc-gen-go.
// source: sfxproto/signalfx.proto
// DO NOT EDIT!

/*
Package sfxproto is a generated protocol buffer package.

It is generated from these files:
	sfxproto/signalfx.proto

It has these top-level messages:
	Datum
	Dimension
	DataPoint
	DataPointUploadMessage
*/
package sfxproto

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type MetricType int32

const (
	// Numerical: Periodic, instantaneous measurement of some state.
	MetricType_GAUGE MetricType = 0
	// Numerical: Count of occurrences. Generally non-negative integers.
	MetricType_COUNTER MetricType = 1
	// String: Used for non-continuous quantities (that is, measurements where
	// there is a fixed set of meaningful values). This is essentially a special
	// case of gauge.
	MetricType_ENUM MetricType = 2
	// Tracks a value that increases over time, where only the difference is
	// important.
	MetricType_CUMULATIVE_COUNTER MetricType = 3
)

var MetricType_name = map[int32]string{
	0: "GAUGE",
	1: "COUNTER",
	2: "ENUM",
	3: "CUMULATIVE_COUNTER",
}
var MetricType_value = map[string]int32{
	"GAUGE":              0,
	"COUNTER":            1,
	"ENUM":               2,
	"CUMULATIVE_COUNTER": 3,
}

func (x MetricType) Enum() *MetricType {
	p := new(MetricType)
	*p = x
	return p
}
func (x MetricType) String() string {
	return proto.EnumName(MetricType_name, int32(x))
}
func (x *MetricType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(MetricType_value, data, "MetricType")
	if err != nil {
		return err
	}
	*x = MetricType(value)
	return nil
}

type Datum struct {
	StrValue         *string  `protobuf:"bytes,1,opt,name=strValue" json:"strValue,omitempty"`
	DoubleValue      *float64 `protobuf:"fixed64,2,opt,name=doubleValue" json:"doubleValue,omitempty"`
	IntValue         *int64   `protobuf:"varint,3,opt,name=intValue" json:"intValue,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Datum) Reset()         { *m = Datum{} }
func (m *Datum) String() string { return proto.CompactTextString(m) }
func (*Datum) ProtoMessage()    {}

func (m *Datum) GetStrValue() string {
	if m != nil && m.StrValue != nil {
		return *m.StrValue
	}
	return ""
}

func (m *Datum) GetDoubleValue() float64 {
	if m != nil && m.DoubleValue != nil {
		return *m.DoubleValue
	}
	return 0
}

func (m *Datum) GetIntValue() int64 {
	if m != nil && m.IntValue != nil {
		return *m.IntValue
	}
	return 0
}

type Dimension struct {
	Key              *string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value            *string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Dimension) Reset()         { *m = Dimension{} }
func (m *Dimension) String() string { return proto.CompactTextString(m) }
func (*Dimension) ProtoMessage()    {}

func (m *Dimension) GetKey() string {
	if m != nil && m.Key != nil {
		return *m.Key
	}
	return ""
}

func (m *Dimension) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

type DataPoint struct {
	// source, field 1, was deprecated, so start at field 2
	Metric           *string      `protobuf:"bytes,2,opt,name=metric" json:"metric,omitempty"`
	Timestamp        *int64       `protobuf:"varint,3,opt,name=timestamp" json:"timestamp,omitempty"`
	Value            *Datum       `protobuf:"bytes,4,opt,name=value" json:"value,omitempty"`
	MetricType       *MetricType  `protobuf:"varint,5,opt,name=metricType,enum=sfxproto.MetricType" json:"metricType,omitempty"`
	Dimensions       []*Dimension `protobuf:"bytes,6,rep,name=dimensions" json:"dimensions,omitempty"`
	XXX_unrecognized []byte       `json:"-"`
}

func (m *DataPoint) Reset()         { *m = DataPoint{} }
func (*DataPoint) ProtoMessage()    {}

func (m *DataPoint) GetMetric() string {
	if m != nil && m.Metric != nil {
		return *m.Metric
	}
	return ""
}

func (m *DataPoint) GetTimestamp() int64 {
	if m != nil && m.Timestamp != nil {
		return *m.Timestamp
	}
	return 0
}

func (m *DataPoint) GetValue() *Datum {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *DataPoint) GetMetricType() MetricType {
	if m != nil && m.MetricType != nil {
		return *m.MetricType
	}
	return MetricType_GAUGE
}

func (m *DataPoint) GetDimensions() []*Dimension {
	if m != nil {
		return m.Dimensions
	}
	return nil
}

type DataPointUploadMessage struct {
	Datapoints       []*DataPoint `protobuf:"bytes,1,rep,name=datapoints" json:"datapoints,omitempty"`
	XXX_unrecognized []byte       `json:"-"`
}

func (m *DataPointUploadMessage) Reset()         { *m = DataPointUploadMessage{} }
func (m *DataPointUploadMessage) String() string { return proto.CompactTextString(m) }
func (*DataPointUploadMessage) ProtoMessage()    {}

func (m *DataPointUploadMessage) GetDatapoints() []*DataPoint {
	if m != nil {
		return m.Datapoints
	}
	return nil
}

func init() {
	proto.RegisterEnum("sfxproto.MetricType", MetricType_name, MetricType_value)
}
