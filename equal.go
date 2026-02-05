package cofly

func Equal(oldValue, newValue any) bool {
	if newValue == nil || oldValue == nil {
		return newValue == oldValue
	}

	switch newValue := newValue.(type) {
	case bool:
		oldValue, ok := oldValue.(bool)
		return ok && newValue == oldValue
	case int:
		switch oldValue := oldValue.(type) {
		case int:
			return newValue == oldValue
		case float64:
			return float64(newValue) == oldValue
		default:
			return false
		}
	case float64:
		switch oldValue := oldValue.(type) {
		case int:
			return newValue == float64(oldValue)
		case float64:
			return newValue == oldValue
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
