package cofly

import "fmt"

func Clone(value any) any {
	if value == nil {
		return nil
	}

	switch value := value.(type) {
	case bool:
		return value
	case int:
		return value
	case float64:
		return value
	case string:
		return value
	case map[string]any:
		return cloneMap(value)
	case []any:
		return cloneArray(value)
	}

	panic(fmt.Sprintf("type [%T] unsupported", value))
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
