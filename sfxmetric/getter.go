package sfxmetric

// Getter is an interface that is used by Metric. It may return any kind of
// integer, float, string or nil, including pointers. Any other type is invalid.
type Getter interface {
	Get() (interface{}, error)
}

type GetterFunc func() (interface{}, error)

func (f GetterFunc) Get() (interface{}, error) {
	return f()
}

// IntValueGetter is a convenience function for making a value satisfy the
// Getter interface. It is especially useful with pointers.
func ValueGetter(value interface{}) Getter {
	return valueGetter{value}
}

type valueGetter struct {
	value interface{}
}

func (v valueGetter) Get() (interface{}, error) {
	return v.value, nil
}
