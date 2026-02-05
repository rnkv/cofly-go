package cofly

import (
	"cmp"
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
	//
}
