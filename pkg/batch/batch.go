package batch

func IsAllTrue[T any](elem []T, predicate func(T) bool) bool {
	for _, e := range elem {
		if !predicate(e) {
			return false
		}
	}
	return true
}
