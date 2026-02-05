package cofly

import (
	"fmt"
)

func Merge(target any, change any, doClean bool) any {
	if target == nil {
		return change
	}

	if change == nil {
		return nil
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

		changeSplices := parseSplices(change)

		if len(changeSplices) > 0 {
			switch target := target.(type) {
			case map[string]any:
				targetSplices := parseSplices(target)

				if len(targetSplices) == 0 {
					panic("target is not splices")
				}

				return mergeSplicesIntoSplices(targetSplices, changeSplices, doClean)
			case []any:
				return mergeSplicesIntoArray(target, changeSplices, doClean)
			default:
				panic(fmt.Sprintf("target type [%T] is not supported", target))
			}
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
			return change
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

func mergeMapIntoMap(targetMap map[string]any, changeMap map[string]any, doClean bool) map[string]any {
	for changeKey, changeValue := range changeMap {
		if changeValue == Undefined {
			if doClean {
				delete(targetMap, changeKey)
				continue
			}

			targetMap[changeKey] = Undefined
			continue
		}

		targetValue, doesTargetValueExist := targetMap[changeKey]
		if doesTargetValueExist {
			targetMap[changeKey] = Merge(targetValue, changeValue, doClean)
		} else {
			targetMap[changeKey] = changeValue
		}
	}

	return targetMap
}

func mergeSplicesIntoArray(targetArray []any, changeSplices []splice, doClean bool) []any {
	sortSplices(changeSplices)
	validateSplices(changeSplices)

	// fmt.Printf("changeSplices: %#v\n", changeSplices)
	// fmt.Printf("targetArrayLength: %d\n", len(targetArray))
	outputArrayLength := len(targetArray)

	for _, changeSplice := range changeSplices {
		insideArraySpan, isInsideArraySpan := changeSplice.span.positiveBeforeLength(len(targetArray))
		outputArrayLength += len(changeSplice.value)

		if isInsideArraySpan {
			outputArrayLength -= insideArraySpan.length()
		}
	}

	// fmt.Printf("outputArrayLength: %d\n", outputArrayLength)
	outputArray := make([]any, 0, outputArrayLength)
	targetArrayCursor := 0
	changeSpliceValueOffset := 0
	changeSpliceIndex := 0

	for changeSpliceIndex < len(changeSplices) {
		changeSplice := changeSplices[changeSpliceIndex]
		beforeArraySpan, isBeforeArraySpan := changeSplice.span.negative()
		if !isBeforeArraySpan {
			break
		}

		// fmt.Printf("beforeArraySpan: %#v\n", beforeArraySpan)
		_, isInsideArraySpan := changeSplice.span.positiveBeforeLength(len(targetArray))
		appendedElementsCount := 0

		if isInsideArraySpan {
			appendedElementsCount = min(beforeArraySpan.length(), len(changeSplice.value))
		} else {
			appendedElementsCount = len(changeSplice.value)
		}

		// fmt.Printf("appendedElementsCount: %d\n", appendedElementsCount)

		outputArray = append(
			outputArray,
			changeSplice.value[0:appendedElementsCount]...,
		)

		if isInsideArraySpan {
			changeSpliceValueOffset += appendedElementsCount
			break
		}

		changeSpliceIndex++
	}

	for changeSpliceIndex < len(changeSplices) {
		changeSplice := changeSplices[changeSpliceIndex]
		insideArraySpan, isInsideArraySpan := changeSplice.span.positiveBeforeLength(len(targetArray))
		if !isInsideArraySpan {
			changeSpliceValueOffset = 0
			break
		}

		// fmt.Printf("insideArraySpan: %#v\n", insideArraySpan)
		outputArray = append(outputArray, targetArray[targetArrayCursor:insideArraySpan.indexFrom]...)
		targetArrayCursor = insideArraySpan.indexTo

		modifiedElementsCount := min(
			insideArraySpan.length(),
			len(changeSplice.value)-changeSpliceValueOffset,
		)

		// fmt.Printf("modifiedElementsCount: %d\n", modifiedElementsCount)

		for elementIndex := range modifiedElementsCount {
			outputArray = append(outputArray, Merge(
				targetArray[insideArraySpan.indexFrom+elementIndex],
				changeSplice.value[changeSpliceValueOffset+elementIndex],
				doClean,
			))
		}

		changeSpliceValueOffset += modifiedElementsCount
		appendedElementsCount := max(0, len(changeSplice.value)-changeSpliceValueOffset)
		// fmt.Printf("appendedElementsCount: %d\n", appendedElementsCount)

		outputArray = append(
			outputArray,
			changeSplice.value[changeSpliceValueOffset:changeSpliceValueOffset+appendedElementsCount]...,
		)

		_, isAfterArraySpan := changeSplice.span.afterLength(len(targetArray))
		if isAfterArraySpan {
			changeSpliceValueOffset += appendedElementsCount
			break
		}

		changeSpliceValueOffset = 0
		changeSpliceIndex++
	}

	if targetArrayCursor < len(targetArray) {
		outputArray = append(outputArray, targetArray[targetArrayCursor:]...)
		targetArrayCursor = len(targetArray)
	}

	for changeSpliceIndex < len(changeSplices) {
		changeSplice := changeSplices[changeSpliceIndex]
		_, isAfterArraySpan := changeSplice.span.afterLength(len(targetArray))
		if !isAfterArraySpan {
			panic("impossible case")
		}

		// fmt.Printf("afterArraySpan: %#v\n", afterArraySpan)
		appendedElementsCount := max(0, len(changeSplice.value)-changeSpliceValueOffset)

		outputArray = append(
			outputArray,
			changeSplice.value[changeSpliceValueOffset:changeSpliceValueOffset+appendedElementsCount]...,
		)

		changeSpliceValueOffset = 0
		changeSpliceIndex++
	}

	if outputArrayLength != len(outputArray) {
		panic(fmt.Sprintf("outputArrayLength: %d, len(outputArray): %d", outputArrayLength, len(outputArray)))
	}

	return outputArray
}

func mergeSplicesIntoSplices(
	targetSplices []splice,
	changeSplices []splice,
	doClean bool,
) any {
	panic("not implemented")
}
