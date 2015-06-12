package sfxproto

import (
	"strings"
	"unicode"
)

// Dimensions is map that can be converted into []*Dimension
type Dimensions map[string]string

func (ds Dimensions) List() []*Dimension {
	ret := make([]*Dimension, 0, len(ds))

	for key, val := range ds {
		if key == "" || val == "" {
			continue
		}

		ret = append(ret, &Dimension{
			Key:   massageKey(key),
			Value: val,
		})
	}

	return ret
}

func (ds Dimensions) Concat(val Dimensions) Dimensions {
	ret := make(Dimensions, len(ds)+len(val))

	for key, val := range ds {
		ret[key] = val
	}

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

func (ds Dimensions) Clone() Dimensions {
	ret := Dimensions{}
	for key, val := range ds {
		ret[key] = val
	}
	return ret
}

func NewDimensions(dims []*Dimension) Dimensions {
	ret := make(Dimensions, len(dims))

	for _, dim := range dims {
		ret[dim.Key] = dim.Value
	}

	return ret
}
