package signalfx

func toInt64(val interface{}) (int64, error) {
	switch tval := val.(type) {
	case int:
		return int64(tval), nil
	case *int:
		return int64(*tval), nil
	case int8:
		return int64(tval), nil
	case *int8:
		return int64(*tval), nil
	case int16:
		return int64(tval), nil
	case *int16:
		return int64(*tval), nil
	case int32:
		return int64(tval), nil
	case *int32:
		return int64(*tval), nil
	case int64:
		return tval, nil
	case *int64:
		return *tval, nil
	case uint:
		return int64(tval), nil
	case *uint:
		return int64(*tval), nil
	case uint8:
		return int64(tval), nil
	case *uint8:
		return int64(*tval), nil
	case uint16:
		return int64(tval), nil
	case *uint16:
		return int64(*tval), nil
	case uint32:
		return int64(tval), nil
	case *uint32:
		return int64(*tval), nil
	case uint64:
		return int64(tval), nil
	case *uint64:
		return int64(*tval), nil
	default:
		return 0, ErrIllegalType
	}
}

func toFloat64(val interface{}) (float64, error) {
	switch tval := val.(type) {
	case float32:
		return float64(tval), nil
	case *float32:
		return float64(*tval), nil
	case float64:
		return tval, nil
	case *float64:
		return *tval, nil
	default:
		return 0, ErrIllegalType
	}
}

func toString(val interface{}) (string, error) {
	switch tval := val.(type) {
	case string:
		return tval, nil
	case *string:
		return *tval, nil
	default:
		return "", ErrIllegalType
	}
}
