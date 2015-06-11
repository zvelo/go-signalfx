package sfxproto

// NewDimension creates a new Dimension with the given key and value
func NewDimension(key, value string) *Dimension {
	return &Dimension{
		Key:   key,
		Value: value,
	}
}

// Dimensions is a Dimension list
type Dimensions map[string]string

func (ds Dimensions) List() []*Dimension {
	ret := make([]*Dimension, 0, len(ds))

	for key, val := range ds {
		if key == "" || val == "" {
			continue
		}

		ret = append(ret, NewDimension(massageKey(key), val))
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
