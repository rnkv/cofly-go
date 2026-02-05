package cofly

import (
	"fmt"
)

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
			return mapDifference(oldValue, newValue)
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
			return arrayDifference(oldValue, newValue)
		default:
			panic(fmt.Sprintf("type [%T] unsupported", oldValue))
		}
	default:
		panic(fmt.Sprintf("type [%T] unsupported", newValue))
	}
}

// mapDifference is a helper function that calculates the difference between two maps
func mapDifference(oldMap, newMap map[string]any) any {
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

func arrayDifference(oldArray, newArray []any) any {
	type operation int

	const (
		skip operation = iota
		insert
		delete
	)

	n, m := len(oldArray), len(newArray)
	max := n + m

	if max == 0 {
		return Undefined
	}

	// Fast paths (keep exact output contract).
	if n == 0 {
		changes := make(map[string]any, 1)
		changes[newSpan(0, 0).string()] = newArray
		return changes
	}

	if m == 0 {
		changes := make(map[string]any, 1)
		changes[newSpan(0, n).string()] = make([]any, 0)
		return changes
	}

	// Exact equality fast path (avoids trace/operations allocations for common no-op updates).
	if n == m {
		allEqual := true
		for index := range n {
			if !Equal(oldArray[index], newArray[index]) {
				allEqual = false
				break
			}
		}
		if allEqual {
			return Undefined
		}
	}

	offset := max
	v := make([]int, 2*max+1)

	for i := range v {
		v[i] = -1
	}

	v[offset+1] = 0
	trace := make([][]int, 0, max+1)

	// Reduce GC pressure by storing all trace rows in one buffer when reasonably sized.
	// For max <= 2000 (your stated upper bound per array is ~1000), this is ~2M ints.
	totalTraceInts := int64(max+1) * int64(max+2) / 2
	useTraceArena := totalTraceInts > 0 && totalTraceInts <= 4_000_000
	var traceArena []int
	if useTraceArena {
		traceArena = make([]int, 0, int(totalTraceInts))
	}

	appendTrace := func(d int) {
		rowLen := d + 1
		var row []int

		if useTraceArena {
			start := len(traceArena)
			traceArena = traceArena[:start+rowLen]
			row = traceArena[start : start+rowLen]
		} else {
			row = make([]int, rowLen)
		}

		for k := -d; k <= d; k += 2 {
			row[(k+d)/2] = v[offset+k]
		}

		trace = append(trace, row)
	}

	for d := 0; d <= max; d++ {
		for k := -d; k <= d; k += 2 {
			kIndex := offset + k
			var x int

			if k == -d || (k != d && v[kIndex-1] < v[kIndex+1]) {
				// insertion
				x = v[kIndex+1]
			} else {
				// deletion
				x = v[kIndex-1] + 1
			}

			y := x - k

			for x < n && y < m && Equal(oldArray[x], newArray[y]) {
				x++
				y++
			}

			v[kIndex] = x

			if x >= n && y >= m {
				// store final v for this d (compact form)
				appendTrace(d)
				goto BACKTRACK
			}
		}

		appendTrace(d)
	}

BACKTRACK:
	getTrace := func(row []int, k int) int {
		// For a given d, trace row stores x values for k in [-d..d] with step=2.
		d := len(row) - 1
		return row[(k+d)/2]
	}

	operations := make([]operation, 0, max)
	x, y := n, m

	for d := len(trace) - 1; d > 0; d-- {
		vPrev := trace[d-1]
		k := x - y

		var prevK int
		if k == -d || (k != d && getTrace(vPrev, k-1) < getTrace(vPrev, k+1)) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := getTrace(vPrev, prevK)
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			operations = append(operations, skip)
			x--
			y--
		}

		if x == prevX {
			operations = append(operations, insert)
			y--
		} else {
			operations = append(operations, delete)
			x--
		}
	}

	// tail of the snake at d=0
	for x > 0 && y > 0 {
		operations = append(operations, skip)
		x--
		y--
	}

	// operations are in reverse order
	for i, j := 0, len(operations)-1; i < j; i, j = i+1, j-1 {
		operations[i], operations[j] = operations[j], operations[i]
	}

	changes := make(map[string]any)
	oldI, newI := 0, 0
	open := false
	curFrom, curTo := 0, 0
	curValue := make([]any, 0)

	flush := func() {
		if !open {
			return
		}

		// For replacements (delete+insert in the same splice), store element-level diffs
		// instead of full new values. This makes the resulting patch "speak" the same
		// language as Merge(): it will Merge(oldElem, diffElem) to reach newElem.
		//
		// For i in [0..rep), we are replacing oldArray[curFrom+i] with curValue[i].
		// For i >= rep, curValue is a pure insertion payload.
		delLen := curTo - curFrom
		replacementsCount := min(delLen, len(curValue))

		for i := range replacementsCount {
			change := Difference(oldArray[curFrom+i], curValue[i])

			if change == Undefined {
				// Should be rare (Myers should align equal elements), but never emit the
				// Undefined marker as a change value, because Merge() treats it specially.
				curValue[i] = oldArray[curFrom+i]
			} else {
				curValue[i] = change
			}
		}

		span := span{indexFrom: curFrom, indexTo: curTo}
		changes[span.string()] = curValue
		open = false
		curValue = make([]any, 0)
	}

	for _, operation := range operations {
		switch operation {
		case skip:
			flush()
			oldI++
			newI++
		case delete:
			if !open {
				open = true
				curFrom = oldI
				curTo = oldI
				curValue = make([]any, 0)
			}

			curTo++
			oldI++
		case insert:
			if !open {
				open = true
				curFrom = oldI
				curTo = oldI
				curValue = make([]any, 0)
			}
			curValue = append(curValue, newArray[newI])
			newI++
		}
	}

	flush()

	if len(changes) == 0 {
		return Undefined
	}

	return changes
}

// func arrayDifference__OLD(oldArray, newArray []any, offset int) any {
// 	oldArrayLength := len(oldArray)
// 	newArrayLength := len(newArray)
// 	maxArrayLength := max(oldArrayLength, newArrayLength)
// 	changes := make(map[string]any, maxArrayLength)

// 	for index := range maxArrayLength {
// 		doesOldValueExist := index < oldArrayLength
// 		doesNewValueExist := index < newArrayLength

// 		if doesOldValueExist && doesNewValueExist {
// 			change := Difference(oldArray[index], newArray[index])
// 			if change != Undefined {
// 				changes[strconv.FormatInt(int64(index+offset), 10)] = change
// 			}
// 		} else if doesOldValueExist {
// 			changes[strconv.FormatInt(int64(index+offset), 10)] = Undefined
// 		} else if doesNewValueExist {
// 			changes[strconv.FormatInt(int64(index+offset), 10)] = newArray[index]
// 		} else {
// 			panic("impossible case")
// 		}
// 	}

// 	if len(changes) == 0 {
// 		return Undefined
// 	}

// 	return changes
// }

// func arrayDifferenceWithOverlap(oldArray, newArray []any) any {
// 	overlappedElementsCount, oldArrayOverlapOffset, newArrayOverlapOffset := overlap(oldArray, newArray)

// 	if overlappedElementsCount == 0 {
// 		return arrayDifference__OLD(oldArray, newArray, 0)
// 	}

// 	var prefixChangeMap map[string]any
// 	var suffixChangeMap map[string]any

// 	if oldArrayOverlapOffset > 0 || newArrayOverlapOffset > 0 {
// 		oldArrayPrefix := oldArray[:oldArrayOverlapOffset]
// 		newArrayPrefix := newArray[:newArrayOverlapOffset]
// 		elementsCountDifference := len(oldArrayPrefix) - len(newArrayPrefix)
// 		maxPrefixLength := max(len(oldArrayPrefix), len(newArrayPrefix))

// 		prefixChangeMap = make(map[string]any, maxPrefixLength)
// 		removedElementsCount := max(0, elementsCountDifference)
// 		addedElementsCount := max(0, -elementsCountDifference)

// 		for index := range removedElementsCount {
// 			prefixChangeMap[strconv.FormatInt(int64(index), 10)] = Undefined
// 		}

// 		prefixBodyChange := arrayDifference__OLD(
// 			oldArrayPrefix[removedElementsCount:],
// 			newArrayPrefix[addedElementsCount:],
// 			removedElementsCount,
// 		)

// 		if prefixBodyChange != Undefined {
// 			maps.Copy(prefixChangeMap, prefixBodyChange.(map[string]any))
// 		}

// 		for index := range addedElementsCount {
// 			prefixChangeMap[strconv.FormatInt(int64(-addedElementsCount+index), 10)] = newArrayPrefix[index]
// 		}
// 	}

// 	oldArraySuffix := oldArray[oldArrayOverlapOffset+overlappedElementsCount:]
// 	newArraySuffix := newArray[newArrayOverlapOffset+overlappedElementsCount:]
// 	suffixChange := arrayDifference__OLD(oldArraySuffix, newArraySuffix, oldArrayOverlapOffset+overlappedElementsCount)

// 	if suffixChange != Undefined {
// 		suffixChangeMap = suffixChange.(map[string]any)
// 	}

// 	if prefixChangeMap == nil && suffixChangeMap == nil {
// 		return Undefined
// 	}

// 	changeMap := make(map[string]any, len(prefixChangeMap)+len(suffixChangeMap))
// 	maps.Copy(changeMap, prefixChangeMap)
// 	maps.Copy(changeMap, suffixChangeMap)
// 	return changeMap
// }
