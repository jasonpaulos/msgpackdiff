package msgpackdiff

// Compare checks two MessagePack objects for equality. The first return value will be true if and
// only if the objects a and b are considered equivalent. If the second return value is a non-nil
// error, then the comparison could not be completed and the first return value should be ignored.
func Compare(a []byte, b []byte, stopOnFirstDifference bool, ignoreEmpty bool, ignoreOrder bool) (bool, error) {
	objA, _, errA := Parse(a)
	if errA != nil {
		return false, errA
	}

	objB, _, errB := Parse(b)
	if errB != nil {
		return false, errB
	}

	// TODO: actually compare objects
	return objA == objB, nil
}
