package sfxmetric

// Getter is an interface that is used by Metric. It may return any kind of
// integer, float, string or nil, including pointers. Any other type is invalid.
type Getter interface {
	Get() (interface{}, error)
}

// The GetterFunc type is an adapter to allow the use of ordinary functions as
// Metric Getters. If f is a function with the appropriate signature,
// GetterFunc(f) is a Getter object that calls f.
type GetterFunc func() (interface{}, error)

// Get calls f()
func (f GetterFunc) Get() (interface{}, error) {
	return f()
}

// ValueGetter is a convenience function for making a value satisfy the
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
