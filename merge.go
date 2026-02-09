package cofly

import (
	"fmt"
)

func Merge(target any, change any, doClean bool) any {
	switch change := change.(type) {
	case nil, bool, int, float64:
		return change
	case string:
		if change == Undefined {
			return target
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
		case map[string]any:
			return mergeMapIntoMap(target, change, doClean)
		case nil, bool, int, float64, string, []any:
			return change
		default:
			panic(fmt.Sprintf("target type [%T] is not supported", target))
		}
	case []any:
		if change == nil {
			return nil
		}

		switch target.(type) {
		case nil, bool, int, float64, string, map[string]any, []any:
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
		if changeSplice.span.indexTo > len(targetArray) {
			panic(fmt.Sprintf("changeSplice span is past the end of the array: %q", changeSplice.span.string()))
		}

		outputArrayLength += len(changeSplice.value) - changeSplice.span.length()
	}

	// fmt.Printf("outputArrayLength: %d\n", outputArrayLength)
	outputArray := make([]any, 0, outputArrayLength)
	targetArrayCursor := 0
	changeSpliceValueOffset := 0

	for _, changeSplice := range changeSplices {
		outputArray = append(outputArray, targetArray[targetArrayCursor:changeSplice.span.indexFrom]...)
		targetArrayCursor = changeSplice.span.indexTo

		modifiedElementsCount := min(
			changeSplice.span.length(),
			len(changeSplice.value)-changeSpliceValueOffset,
		)

		// fmt.Printf("modifiedElementsCount: %d\n", modifiedElementsCount)

		for elementIndex := range modifiedElementsCount {
			outputArray = append(outputArray, Merge(
				targetArray[changeSplice.span.indexFrom+elementIndex],
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

		changeSpliceValueOffset = 0
	}

	if targetArrayCursor < len(targetArray) {
		outputArray = append(outputArray, targetArray[targetArrayCursor:]...)
		targetArrayCursor = len(targetArray)
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
