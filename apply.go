package cofly

func Apply(target *any, isSnapshot bool, change *any, doClean bool) bool {
	if !isSnapshot {
		*target = Merge(*target, *change, doClean)
		return true
	}

	snapshot := *change
	*change = Difference(*target, snapshot)

	if *change == Undefined {
		return false
	}

	*target = snapshot
	return true
}
