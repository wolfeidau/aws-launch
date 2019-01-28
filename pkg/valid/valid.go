package valid

// OneOf returns true if is there one not nil value in the list of values
func OneOf(values ...interface{}) bool {
	return CountOfNotNil(values...) == 1
}

// CountOfNotNil count of values which omits nils
func CountOfNotNil(values ...interface{}) int {
	c := 0
	for _, v := range values {
		if v != nil {
			c++
		}
	}

	return c
}
