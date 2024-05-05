package utils

func AreMutuallyExclusive(values ...string) bool {
	var nonNullValues []string
	for _, v := range values {
		if v != "" {
			nonNullValues = append(nonNullValues, v)
		}
	}

	return len(nonNullValues) == 1
}
