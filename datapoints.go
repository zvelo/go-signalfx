package signalfx

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// DataPoints represents a set of DataPoint objects
type DataPoints struct {
	datapoints map[*DataPoint]interface{}
	mu         sync.Mutex
}

// NewDataPoints returns a new DataPoints object with an expected, length, of l
func NewDataPoints(l int) *DataPoints {
	return &DataPoints{
		datapoints: make(map[*DataPoint]interface{}, l),
	}
}

func (dps *DataPoints) lock() {
	dps.mu.Lock()
}

func (dps *DataPoints) unlock() {
	dps.mu.Unlock()
}

// Add a DataPoint to the set
func (dps *DataPoints) Add(vals ...*DataPoint) *DataPoints {
	dps.lock()
	defer dps.unlock()

	for _, dp := range vals {
		if dp != nil {
			dps.datapoints[dp] = nil
		}
	}

	return dps
}

// Append appends the passed datapoints to the source object
func (dps *DataPoints) Append(val *DataPoints) {
	if val == nil {
		return
	}

	val.lock()
	defer val.unlock()

	for dp := range val.datapoints {
		dps.Add(dp)
	}
}

// List returns a slice of a copy of all of the datapoints contained in DataPoints
func (dps *DataPoints) List() []*DataPoint {
	ret := make([]*DataPoint, 0, dps.Len())

	dps.lock()
	defer dps.unlock()

	for dp := range dps.datapoints {
		ret = append(ret, dp.Clone())
	}

	return ret
}

// Clone makes a copy of the DataPoints. It copies the underlying DataPoint
// pointers and does not do a deep copy of their values.
func (dps *DataPoints) Clone() *DataPoints {
	ret := &DataPoints{
		datapoints: make(map[*DataPoint]interface{}, dps.Len()),
	}

	dps.lock()
	defer dps.unlock()

	for dp := range dps.datapoints {
		ret.datapoints[dp] = nil
	}

	return ret
}

// Remove DataPoint(s) from the set. The match is by testing for pointer
// equality, not DataPoint equality.
func (dps *DataPoints) Remove(vals ...*DataPoint) {
	dps.lock()
	defer dps.unlock()

	for _, dp := range vals {
		delete(dps.datapoints, dp)
	}
}

// RemoveDataPoints removes DataPoint(s) from the set
func (dps *DataPoints) RemoveDataPoints(val *DataPoints) {
	val.lock()
	defer val.unlock()

	for dp := range val.datapoints {
		dps.Remove(dp)
	}
}

// ProtoDataPoints returns a sfxproto.DataPoints object representing the
// underlying DataPoint objects contained in the DataPoints object.
func (dps *DataPoints) ProtoDataPoints() (*sfxproto.DataPoints, error) {
	ret := sfxproto.NewDataPoints(dps.Len())

	dps.lock()
	defer dps.unlock()

	for dp := range dps.datapoints {

		ret.Add(dp.pdp)
	}

	return ret, nil
}

// Len returns the number of datapoints the DataPoints object contains
func (dps *DataPoints) Len() int {
	dps.lock()
	defer dps.unlock()

	return len(dps.datapoints)
}

// filter returns a new DataPoints structure consisting of only those
// points where filterFunc is true
func (dps *DataPoints) filter(filterFunc func(*DataPoint) bool) *DataPoints {
	ret := &DataPoints{
		datapoints: make(map[*DataPoint]interface{}, dps.Len()),
	}

	dps.lock()
	defer dps.unlock()

	for dp := range dps.datapoints {
		if filterFunc(dp) {
			ret.datapoints[dp] = nil
		}
	}

	return ret
}
