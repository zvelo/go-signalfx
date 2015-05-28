package sfxproto

import "github.com/gogo/protobuf/proto"

// NewDimension creates a new Dimension with the given key and value
func NewDimension(key, value string) *Dimension {
	return &Dimension{
		Key:   key,
		Value: value,
	}
}

// Dimensions is a Dimension list
type Dimensions []*Dimension

// Clone makes a deep copy of Dimensions
func (ds Dimensions) Clone() Dimensions {
	ret := make(Dimensions, len(ds))

	for i, d := range ds {
		ret[i] = proto.Clone(d).(*Dimension)
	}

	return ret
}
