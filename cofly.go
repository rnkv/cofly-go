package cofly

import (
	"fmt"
	"slices"
	"strconv"
)

const Undefined string = "\x00"

func Merge(target any, change any, doClean bool) any {
	if target == nil {
		return change
	}

	switch change := change.(type) {
	case bool:
		return change
	case int:
		return change
	case float64:
		return change
	case string:
		if change == Undefined {
			panic("can't merge undefined value")
		}

		return change
	case map[string]any:
		if change == nil {
			return nil
		}

		switch target := target.(type) {
		case bool:
			return change
		case int:
			return change
		case float64:
			return change
		case string:
			return change
		case map[string]any:
			return mergeMapIntoMap(target, change, doClean)
		case []any:
			return mergeMapIntoArray(target, change, doClean)
		default:
			panic(fmt.Sprintf("target type [%T] is not supported", target))
		}
	case []any:
		if change == nil {
			return nil
		}

		switch target.(type) {
		case bool:
			return change
		case int:
			return change
		case float64:
			return change
		case string:
			return change
		case map[string]any:
			return change
		case []any:
			return change
		default:
			panic(fmt.Sprintf("target type [%T] is not supported", target))
		}
	}

	panic(fmt.Sprintf("change type [%T] is not supported", change))
}

func mergeMapIntoMap(targetMap map[string]any, sourceMap map[string]any, doClean bool) map[string]any {
	for sourceKey, changeValue := range sourceMap {
		if changeValue == Undefined {
			if doClean {
				delete(targetMap, sourceKey)
				continue
			}

			targetMap[sourceKey] = Undefined
			continue
		}

		targetValue, doesTargetValueExist := targetMap[sourceKey]
		if doesTargetValueExist {
			targetMap[sourceKey] = Merge(targetValue, changeValue, doClean)
		} else {
			targetMap[sourceKey] = changeValue
		}
	}

	return targetMap
}

func mergeMapIntoArray(targetArray []any, sourceMap map[string]any, doClean bool) []any {
	sourceKeys := make([]int64, 0, len(sourceMap))

	for sourceKey := range sourceMap {
		sourceIndex, err := strconv.ParseInt(sourceKey, 10, 64)
		if err != nil {
			continue
		}

		sourceKeys = append(sourceKeys, sourceIndex)
	}

	slices.Sort(sourceKeys)
	removedTargetIndexes := make([]int64, 0, len(sourceKeys))
	var prependedTargetElements []any
	var appendedTargetElements []any

	for _, sourceKey := range sourceKeys {
		changeValue := sourceMap[strconv.FormatInt(sourceKey, 10)]

		switch {
		case sourceKey < 0:
			if changeValue != Undefined || !doClean {
				prependedTargetElements = append(prependedTargetElements, changeValue)
			}
		case sourceKey < int64(len(targetArray)):
			if changeValue == Undefined {
				if doClean {
					removedTargetIndexes = append(removedTargetIndexes, sourceKey)
				} else {
					targetArray[sourceKey] = Undefined
				}
			} else {
				targetArray[sourceKey] = Merge(targetArray[sourceKey], changeValue, doClean)
			}
		default:
			if changeValue != Undefined || !doClean {
				appendedTargetElements = append(appendedTargetElements, changeValue)
			}
		}
	}

	if len(removedTargetIndexes) > 0 {
		targetIndex := 0
		removedTargetIndexIndex := 0

		for sourceIndex := int64(0); sourceIndex < int64(len(targetArray)); sourceIndex++ {
			if removedTargetIndexIndex < len(removedTargetIndexes) &&
				sourceIndex == removedTargetIndexes[removedTargetIndexIndex] {
				removedTargetIndexIndex++
				continue
			}

			targetArray[targetIndex] = targetArray[sourceIndex]
			targetIndex++
		}

		targetArray = targetArray[:targetIndex]
	}

	if len(prependedTargetElements) > 0 {
		targetArray = append(prependedTargetElements, targetArray...)
	}

	if len(appendedTargetElements) > 0 {
		targetArray = append(targetArray, appendedTargetElements...)
	}

	return targetArray
}

func Difference(oldValue any, newValue any) any {
	if newValue == nil {
		if oldValue == nil {
			return Undefined
		}

		return nil
	}

	if oldValue == nil {
		return newValue
	}

	switch newValue := newValue.(type) {
	case bool:
		switch oldValue := oldValue.(type) {
		case bool:
			if newValue == oldValue {
				return Undefined
			}

			return newValue
		case int:
			return newValue
		case float64:
			return newValue
		case string:
			return newValue
		case map[string]any:
			return newValue
		case []any:
			return newValue
		}

		panic(fmt.Sprintf("type [%T] unsupported", oldValue))
	case int:
		switch oldValue := oldValue.(type) {
		case bool:
			return newValue
		case int:
			if newValue == oldValue {
				return Undefined
			}

			return newValue
		case float64:
			if float64(newValue) == oldValue {
				return Undefined
			}

			return newValue
		case string:
			return newValue
		case map[string]any:
			return newValue
		case []any:
			return newValue
		}

		panic(fmt.Sprintf("type [%T] unsupported", oldValue))
	case float64:
		switch oldValue := oldValue.(type) {
		case bool:
			return newValue
		case int:
			if newValue == float64(oldValue) {
				return Undefined
			}

			return newValue
		case float64:
			if newValue == oldValue {
				return Undefined
			}

			return newValue
		case string:
			return newValue
		case map[string]any:
			return newValue
		case []any:
			return newValue
		}

		panic(fmt.Sprintf("type [%T] unsupported", oldValue))
	case string:
		switch oldValue := oldValue.(type) {
		case bool:
			return newValue
		case int:
			return newValue
		case float64:
			return newValue
		case string:
			if newValue == oldValue {
				return Undefined
			}

			return newValue
		case map[string]any:
			return newValue
		case []any:
			return newValue
		}

		panic(fmt.Sprintf("type [%T] unsupported", oldValue))
	case map[string]any:
		switch oldValue := oldValue.(type) {
		case bool:
			return newValue
		case int:
			return newValue
		case float64:
			return newValue
		case string:
			return newValue
		case map[string]any:
			return differenceBetweenMaps(oldValue, newValue)
		case []any:
			panic("type []any unsupported to compare with map[string]any")
		default:
			panic(fmt.Sprintf("type [%T] unsupported", oldValue))
		}
	case []any:
		switch oldValue := oldValue.(type) {
		case bool:
			return newValue
		case int:
			return newValue
		case float64:
			return newValue
		case string:
			return newValue
		case map[string]any:
			return newValue
		case []any:
			return differenceBetweenArrays(oldValue, newValue)
		default:
			panic(fmt.Sprintf("type [%T] unsupported", oldValue))
		}
	default:
		panic(fmt.Sprintf("type [%T] unsupported", newValue))
	}

}

func differenceBetweenMaps(oldMap, newMap map[string]any) any {
	keys := make(map[string]struct{})

	for oldKey := range oldMap {
		keys[oldKey] = struct{}{}
	}

	for newKey := range newMap {
		keys[newKey] = struct{}{}
	}

	changes := make(map[string]any)

	for key := range keys {
		oldValue, doesOldKeyExist := oldMap[key]
		newValue, doesNewKeyExist := newMap[key]

		if doesOldKeyExist && doesNewKeyExist {
			change := Difference(oldValue, newValue)

			if change != Undefined {
				changes[key] = change
			}
		} else if doesOldKeyExist {
			changes[key] = Undefined
		} else if doesNewKeyExist {
			changes[key] = newValue
		} else {
			panic("impossible case")
		}
	}

	if len(changes) == 0 {
		return Undefined
	}

	return changes
}

func differenceBetweenArrays(oldArray, newArray []any) any {
	oldArrayLength := len(oldArray)
	newArrayLength := len(newArray)
	maxArrayLength := max(oldArrayLength, newArrayLength)
	changes := make(map[string]any)

	for index := range maxArrayLength {
		doesOldValueExist := index < oldArrayLength
		doesNewValueExist := index < newArrayLength

		if doesOldValueExist && doesNewValueExist {
			change := Difference(oldArray[index], newArray[index])
			if change != Undefined {
				changes[strconv.FormatUint(uint64(index), 10)] = change
			}
		} else if doesOldValueExist {
			changes[strconv.FormatInt(int64(index), 10)] = Undefined
		} else if doesNewValueExist {
			changes[strconv.FormatInt(int64(index), 10)] = newArray[index]
		} else {
			panic("impossible case")
		}
	}

	if len(changes) == 0 {
		return Undefined
	}

	return changes
}

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

func Apply(target *any, isSnapshot bool, change *any, doClean bool) bool {
	if isSnapshot {
		snapshot := *change
		*change = Difference(*target, snapshot)

		if *change == Undefined {
			return false
		}

		*target = snapshot
	} else {
		*target = Merge(*target, *change, doClean)
	}

	return true
}
