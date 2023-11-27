package utils

func Contains[T comparable](ary []T, item T) bool {
	for _, i := range ary {
		if i == item {
			return true
		}
	}
	return false
}
