package cofly

import "fmt"

func Clone(value any) any {
	switch value := value.(type) {
	case nil, bool, int, float64, string:
		return value
	case map[string]any:
		return cloneMap(value)
	case []any:
		return cloneArray(value)
	default:
		panic(fmt.Sprintf("type [%T] unsupported", value))
	}
}

func cloneMap(sourceMap map[string]any) map[string]any {
	clonedMap := make(map[string]any, len(sourceMap))

	for key, value := range sourceMap {
		clonedMap[key] = Clone(value)
	}

	return clonedMap
}

func cloneArray(sourceArray []any) []any {
	clonedArray := make([]any, len(sourceArray))

	for index, value := range sourceArray {
		clonedArray[index] = Clone(value)
	}

	return clonedArray
}
