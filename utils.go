package cofly

import "fmt"

func toInt64(value any) int64 {
	switch value := value.(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	default:
		panic(fmt.Sprintf("type [%T] unsupported", value))
	}
}

func toUint64(value any) uint64 {
	switch value := value.(type) {
	case uint:
		return uint64(value)
	case uint8:
		return uint64(value)
	case uint16:
		return uint64(value)
	case uint32:
		return uint64(value)
	case uint64:
		return value
	default:
		panic(fmt.Sprintf("type [%T] unsupported", value))
	}
}

func toFloat64(value any) float64 {
	switch value := value.(type) {
	case int:
		return float64(value)
	case int8:
		return float64(value)
	case int16:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case uint:
		return float64(value)
	case uint8:
		return float64(value)
	case uint16:
		return float64(value)
	case uint32:
		return float64(value)
	case uint64:
		return float64(value)
	case float32:
		return float64(value)
	case float64:
		return value
	default:
		panic(fmt.Sprintf("type [%T] unsupported", value))
	}
}
