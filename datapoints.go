package signalfx

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// DataPoints represents a set of DataPoint objects
type DataPoints struct {
	datapoints map[*DataPoint]interface{}
	lock       sync.Mutex
}

// NewDataPoints returns a new DataPoints object with an expected, length, of l
func NewDataPoints(l int) *DataPoints {
	return &DataPoints{
		datapoints: make(map[*DataPoint]interface{}, l),
	}
}

// Add a DataPoint to the set
func (ms *DataPoints) Add(vals ...*DataPoint) *DataPoints {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for _, m := range vals {
		if m != nil {
			ms.datapoints[m] = nil
		}
	}

	return ms
}

// Concat appends the passed datapoints to the source object
func (ms *DataPoints) Concat(val *DataPoints) {
	if val == nil {
		return
	}

	val.lock.Lock()
	defer val.lock.Unlock()

	for m := range val.datapoints {
		ms.Add(m)
	}
}

// List returns a slice of a copy of all of the datapoints contained in DataPoints
func (ms *DataPoints) List() []*DataPoint {
	ret := make([]*DataPoint, 0, ms.Len())

	ms.lock.Lock()
	defer ms.lock.Unlock()

	for m := range ms.datapoints {
		ret = append(ret, m.Clone())
	}

	return ret
}

// Clone makes a copy of the DataPoints. It copies the underlying DataPoint
// pointers and does not do a deep copy of their values.
func (ms *DataPoints) Clone() *DataPoints {
	ret := &DataPoints{
		datapoints: make(map[*DataPoint]interface{}, ms.Len()),
	}

	ms.lock.Lock()
	defer ms.lock.Unlock()

	for m := range ms.datapoints {
		ret.datapoints[m] = nil
	}

	return ret
}

// Remove DataPoint(s) from the set. The match is by testing for pointer
// equality, not DataPoint equality.
func (ms *DataPoints) Remove(vals ...*DataPoint) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for _, m := range vals {
		delete(ms.datapoints, m)
	}
}

// RemoveDataPoints removes DataPoint(s) from the set
func (ms *DataPoints) RemoveDataPoints(val *DataPoints) {
	val.lock.Lock()
	defer val.lock.Unlock()

	for m := range val.datapoints {
		ms.Remove(m)
	}
}

// DataPoints returns a sfxproto.DataPoints object representing the underlying
// DataPoints contained in the DataPoints object. If a DataPoint has a Getter, the
// value will be updated before returning.
func (ms *DataPoints) DataPoints() (*sfxproto.DataPoints, error) {
	ret := sfxproto.NewDataPoints(ms.Len())

	ms.lock.Lock()
	defer ms.lock.Unlock()

	for m := range ms.datapoints {
		if err := m.update(); err != nil {
			return nil, err
		}

		ret.Add(m.dp)
	}

	return ret, nil
}

// Len returns the number of datapoints the DataPoints object contains
func (ms *DataPoints) Len() int {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	return len(ms.datapoints)
}
