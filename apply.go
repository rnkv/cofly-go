package cofly

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
