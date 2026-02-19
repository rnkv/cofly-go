package cofly

func Equal(oldValue, newValue any) bool {
	switch newValue := newValue.(type) {
	case nil:
		return oldValue == nil
	case bool:
		oldValue, ok := oldValue.(bool)
		return ok && newValue == oldValue
	case int, int8, int16, int32, int64:
		switch oldValue := oldValue.(type) {
		case int, int8, int16, int32, int64:
			return toInt64(newValue) == toInt64(oldValue)
		case
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			return toFloat64(newValue) == toFloat64(oldValue)
		default:
			return false
		}
	case uint, uint8, uint16, uint32, uint64:
		switch oldValue := oldValue.(type) {
		case uint, uint8, uint16, uint32, uint64:
			return toUint64(newValue) == toUint64(oldValue)
		case
			int, int8, int16, int32, int64,
			float32, float64:
			return toFloat64(newValue) == toFloat64(oldValue)
		default:
			return false
		}
	case float32, float64:
		switch oldValue := oldValue.(type) {
		case
			float32, float64,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			return toFloat64(newValue) == toFloat64(oldValue)
		default:
			return false
		}
	case string:
		oldValue, ok := oldValue.(string)
		return ok && newValue == oldValue
	case map[string]any:
		oldValue, ok := oldValue.(map[string]any)
		return ok && areMapsEqual(oldValue, newValue)
	case []any:
		oldValue, ok := oldValue.([]any)
		return ok && areArraysEqual(oldValue, newValue)
	default:
		return false
	}
}

func areMapsEqual(oldMap, newMap map[string]any) bool {
	if len(oldMap) != len(newMap) {
		return false
	}

	for key, oldValue := range oldMap {
		newValue, doesNewValueExist := newMap[key]
		if !doesNewValueExist {
			return false
		}

		if !Equal(oldValue, newValue) {
			return false
		}
	}

	return true
}

func areArraysEqual(oldArray, newArray []any) bool {
	if len(oldArray) != len(newArray) {
		return false
	}

	for index, oldValue := range oldArray {
		newValue := newArray[index]

		if !Equal(oldValue, newValue) {
			return false
		}
	}
	return true
}
