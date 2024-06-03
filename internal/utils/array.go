package utils

func IndexOf[T comparable](arr []T, val T) int {
	for pos, v := range arr {
		if v == val {
			return pos
		}
	}
	return -1
}
