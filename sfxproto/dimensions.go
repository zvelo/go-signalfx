package sfxproto

import (
	"strings"
	"unicode"

	"github.com/golang/protobuf/proto"
)

// Dimensions is map that can be converted into []*Dimension. By itself it is
// not goroutine safe.
type Dimensions map[string]string

// List returns a slice of all tracked Dimension objects
func (ds Dimensions) List() []*Dimension {
	ret := make([]*Dimension, 0, len(ds))

	for key, val := range ds {
		if key == "" || val == "" {
			continue
		}

		ret = append(ret, &Dimension{
			Key:   proto.String(massageKey(key)),
			Value: proto.String(val),
		})
	}

	return ret
}

// Append returns a new Dimensions object with the values of both objects
// merged. Keys in val will overwrite equivalent keys in ds.
func (ds Dimensions) Append(val Dimensions) Dimensions {
	ret := ds.Clone()

	for key, val := range val {
		ret[key] = val
	}

	return ret
}

func massageKey(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' {
			return r
		}

		return '_'
	}, str)
}

// Clone makes a copy of the given Dimensions object
func (ds Dimensions) Clone() Dimensions {
	ret := Dimensions{}
	for key, val := range ds {
		ret[key] = val
	}
	return ret
}

// NewDimensions creates a new Dimensions object from a slice of Dimension
// objects
func NewDimensions(dims []*Dimension) Dimensions {
	if dims == nil {
		return Dimensions{}
	}

	ret := make(Dimensions, len(dims))

	for _, dim := range dims {
		if dim.Key != nil && dim.Value != nil {
			ret[*dim.Key] = *dim.Value
		}
	}

	return ret
}

// Equal returns whether or not two sets of dimensions match exactly
func (ds Dimensions) Equal(r Dimensions) bool {
	if len(ds) != len(r) {
		return false
	}

	for key, value := range ds {
		got, ok := r[key]
		if !ok {
			return false
		}

		if got != value {
			return false
		}
	}

	return true
}
