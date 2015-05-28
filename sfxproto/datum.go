package sfxproto

import "fmt"

// NewDatum returns a new Datum object with the value properly set
func NewDatum(val interface{}) *Datum {
	ret := &Datum{}

	if err := ret.Set(val); err != nil {
		return nil
	}

	return ret
}

// Set sets the datum value correctly for all integer, float and string types.
// If any other type is passed in, an error is returned.
func (d *Datum) Set(val interface{}) error {
	d.Reset()

	switch tval := val.(type) {
	case int:
		d.IntValue = int64(tval)
	case int8:
		d.IntValue = int64(tval)
	case int16:
		d.IntValue = int64(tval)
	case int32:
		d.IntValue = int64(tval)
	case int64:
		d.IntValue = tval
	case uint:
		d.IntValue = int64(tval)
	case uint8:
		d.IntValue = int64(tval)
	case uint16:
		d.IntValue = int64(tval)
	case uint32:
		d.IntValue = int64(tval)
	case uint64:
		d.IntValue = int64(tval)
	case float32:
		d.DoubleValue = float64(tval)
	case float64:
		d.DoubleValue = tval
	case string:
		d.StrValue = tval
	default:
		return fmt.Errorf("illegal value type")
	}

	return nil
}
