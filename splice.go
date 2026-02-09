package cofly

import (
	"cmp"
	"fmt"
	"slices"
)

type splice struct {
	span  span
	value []any
}

func parseSplice(key string, value any) (splice, bool) {
	span, ok := parseSpan(key)
	if !ok {
		return splice{}, false
	}

	valueArray, ok := value.([]any)
	if !ok {
		return splice{}, false
	}

	return splice{
		span:  span,
		value: valueArray,
	}, true
}

func parseSplices(changeMap map[string]any) []splice {
	splices := make([]splice, 0, len(changeMap))

	for spliceKey, changeValue := range changeMap {
		splice, ok := parseSplice(spliceKey, changeValue)
		if !ok {
			return nil
		}

		splices = append(splices, splice)
	}

	return splices
}

// func (p arrayPatch) RangeLength() int {
// 	return p.indexTo - p.indexFrom
// }

// func (p arrayPatch) RangeLengthBeforeArray() int {
// 	startIndex := max(p.indexFrom, 0)
// 	endIndex := max(p.indexTo, 0)
// 	return max(0, endIndex-startIndex)
// }

// func (p arrayPatch) RangeLengthInsideArray(array []any) int {
// 	if len(array) <= 0 {
// 		return 0
// 	}

// 	startIndex := max(p.indexFrom, 0)
// 	endIndex := min(p.indexTo, len(array))
// 	return max(0, endIndex-startIndex)
// }

// func (p arrayPatch) RangeLengthAfterArray(array []any) int {
// 	if len(array) <= 0 {
// 		return 0
// 	}

// 	startIndex := min(p.indexFrom, len(array))
// 	endIndex := min(p.indexTo, len(array))
// 	return max(0, endIndex-startIndex)
// }

// func (s splice) RangeString() string {
// 	buffer := make([]byte, 0, 24)
// 	buffer = strconv.AppendInt(buffer, int64(s.span.indexFrom), 10)
// 	buffer = append(buffer, '.', '.')
// 	buffer = strconv.AppendInt(buffer, int64(s.span.indexTo), 10)
// 	return string(buffer)
// }

func sortSplices(splices []splice) {
	slices.SortFunc(splices, func(a, b splice) int {
		return cmp.Compare(a.span.indexFrom, b.span.indexFrom)
	})
}

func validateSplices(splices []splice) {
	if len(splices) <= 1 {
		return
	}

	// Assumes splices are sorted by span.indexFrom.
	previousSpan := splices[0].span

	for i := 1; i < len(splices); i++ {
		currentSpan := splices[i].span

		// Spans are half-open ranges [from, to).
		// Two splices overlap if previousSpan.indexTo > currentSpan.indexFrom.
		if previousSpan.indexTo > currentSpan.indexFrom {
			panic(fmt.Sprintf(
				"invalid splice-map: overlapping spans %q and %q",
				previousSpan.string(),
				currentSpan.string(),
			))
		}

		previousSpan = currentSpan
	}
}
