package metrics

/*
This function checks if PCE peak is near zero shift.
If peak shift is far away, we treat it as not a valid PRNU match.
*/
func IsNearZeroShift(stats PCEStats, maxShift int) bool {
	if maxShift < 0 {
		maxShift = 0
	}
	if absInt(stats.ShiftX) > maxShift {
		return false
	}
	if absInt(stats.ShiftY) > maxShift {
		return false
	}
	return true
}
