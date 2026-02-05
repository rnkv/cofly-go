package cofly

func overlap(a, b []any) (n, aStart, bStart int) {
	bestN, bestA, bestB := 0, 0, 0

	for i := range a {
		for j := range b {
			k := 0

			for i+k < len(a) && j+k < len(b) && Difference(a[i+k], b[j+k]) == Undefined {
				k++
			}

			if k > bestN {
				bestN, bestA, bestB = k, i, j
			}
		}
	}

	return bestN, bestA, bestB
}
