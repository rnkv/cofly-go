package cofly

import (
	"fmt"
	"maps"
	"strconv"
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
	type op int
	const (
		equal op = iota
		ins
		del
	)

	eq := func(a, b any) bool {
		return Difference(a, b) == Undefined
	}

	n, m := len(oldArray), len(newArray)
	max := n + m
	if max == 0 {
		return Undefined
	}

	offset := max
	v := make([]int, 2*max+1)
	for i := range v {
		v[i] = -1
	}
	v[offset+1] = 0

	trace := make([][]int, 0, max+1)

	for d := 0; d <= max; d++ {
		for k := -d; k <= d; k += 2 {
			kIdx := offset + k

			var x int
			if k == -d || (k != d && v[kIdx-1] < v[kIdx+1]) {
				// insertion
				x = v[kIdx+1]
			} else {
				// deletion
				x = v[kIdx-1] + 1
			}
			y := x - k

			for x < n && y < m && eq(oldArray[x], newArray[y]) {
				x++
				y++
			}
			v[kIdx] = x

			if x >= n && y >= m {
				// store final v for this d
				vCopy := make([]int, len(v))
				copy(vCopy, v)
				trace = append(trace, vCopy)
				goto BACKTRACK
			}
		}

		vCopy := make([]int, len(v))
		copy(vCopy, v)
		trace = append(trace, vCopy)
	}

BACKTRACK:
	ops := make([]op, 0, max)
	x, y := n, m

	for d := len(trace) - 1; d > 0; d-- {
		vPrev := trace[d-1]
		k := x - y

		var prevK int
		if k == -d || (k != d && vPrev[offset+k-1] < vPrev[offset+k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := vPrev[offset+prevK]
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			ops = append(ops, equal)
			x--
			y--
		}

		if x == prevX {
			ops = append(ops, ins)
			y--
		} else {
			ops = append(ops, del)
			x--
		}
	}

	// остаток "змейки" на d=0
	for x > 0 && y > 0 {
		ops = append(ops, equal)
		x--
		y--
	}

	// ops сейчас в обратном порядке
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
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
		rep := min(delLen, len(curValue))
		for i := 0; i < rep; i++ {
			d := Difference(oldArray[curFrom+i], curValue[i])
			if d == Undefined {
				// Should be rare (Myers should align equal elements), but never emit the
				// Undefined marker as a change value, because Merge() treats it specially.
				curValue[i] = oldArray[curFrom+i]
			} else {
				curValue[i] = d
			}
		}

		sp := span{indexFrom: curFrom, indexTo: curTo}
		changes[sp.string()] = curValue
		open = false
		curValue = make([]any, 0)
	}

	for _, o := range ops {
		switch o {
		case equal:
			flush()
			oldI++
			newI++
		case del:
			if !open {
				open = true
				curFrom = oldI
				curTo = oldI
				curValue = make([]any, 0)
			}
			curTo++
			oldI++
		case ins:
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

func arrayDifference__OLD(oldArray, newArray []any, offset int) any {
	oldArrayLength := len(oldArray)
	newArrayLength := len(newArray)
	maxArrayLength := max(oldArrayLength, newArrayLength)
	changes := make(map[string]any, maxArrayLength)

	for index := range maxArrayLength {
		doesOldValueExist := index < oldArrayLength
		doesNewValueExist := index < newArrayLength

		if doesOldValueExist && doesNewValueExist {
			change := Difference(oldArray[index], newArray[index])
			if change != Undefined {
				changes[strconv.FormatInt(int64(index+offset), 10)] = change
			}
		} else if doesOldValueExist {
			changes[strconv.FormatInt(int64(index+offset), 10)] = Undefined
		} else if doesNewValueExist {
			changes[strconv.FormatInt(int64(index+offset), 10)] = newArray[index]
		} else {
			panic("impossible case")
		}
	}

	if len(changes) == 0 {
		return Undefined
	}

	return changes
}

func arrayDifferenceWithOverlap(oldArray, newArray []any) any {
	overlappedElementsCount, oldArrayOverlapOffset, newArrayOverlapOffset := overlap(oldArray, newArray)

	if overlappedElementsCount == 0 {
		return arrayDifference__OLD(oldArray, newArray, 0)
	}

	var prefixChangeMap map[string]any
	var suffixChangeMap map[string]any

	if oldArrayOverlapOffset > 0 || newArrayOverlapOffset > 0 {
		oldArrayPrefix := oldArray[:oldArrayOverlapOffset]
		newArrayPrefix := newArray[:newArrayOverlapOffset]
		elementsCountDifference := len(oldArrayPrefix) - len(newArrayPrefix)
		maxPrefixLength := max(len(oldArrayPrefix), len(newArrayPrefix))

		prefixChangeMap = make(map[string]any, maxPrefixLength)
		removedElementsCount := max(0, elementsCountDifference)
		addedElementsCount := max(0, -elementsCountDifference)

		for index := range removedElementsCount {
			prefixChangeMap[strconv.FormatInt(int64(index), 10)] = Undefined
		}

		prefixBodyChange := arrayDifference__OLD(
			oldArrayPrefix[removedElementsCount:],
			newArrayPrefix[addedElementsCount:],
			removedElementsCount,
		)

		if prefixBodyChange != Undefined {
			maps.Copy(prefixChangeMap, prefixBodyChange.(map[string]any))
		}

		for index := range addedElementsCount {
			prefixChangeMap[strconv.FormatInt(int64(-addedElementsCount+index), 10)] = newArrayPrefix[index]
		}
	}

	oldArraySuffix := oldArray[oldArrayOverlapOffset+overlappedElementsCount:]
	newArraySuffix := newArray[newArrayOverlapOffset+overlappedElementsCount:]
	suffixChange := arrayDifference__OLD(oldArraySuffix, newArraySuffix, oldArrayOverlapOffset+overlappedElementsCount)

	if suffixChange != Undefined {
		suffixChangeMap = suffixChange.(map[string]any)
	}

	if prefixChangeMap == nil && suffixChangeMap == nil {
		return Undefined
	}

	changeMap := make(map[string]any, len(prefixChangeMap)+len(suffixChangeMap))
	maps.Copy(changeMap, prefixChangeMap)
	maps.Copy(changeMap, suffixChangeMap)
	return changeMap
}
