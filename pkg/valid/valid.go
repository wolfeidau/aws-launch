package valid

import (
	"github.com/fatih/structs"
)

// ReflectCountOfNotZero count of struct values which omits zero values for the fields with the supplied names
func ReflectCountOfNotZero(v interface{}, names ...string) int {
	if !structs.IsStruct(v) {
		panic("v is not a struct")
	}

	matches := []string{}
	s := structs.New(v)

	// loop over names as it is most likely the shortest array
	for _, name := range names {
		// if a field is not found by name skip it
		if f, ok := s.FieldOk(name); ok {
			// only check exported fields for non zero values
			if f.IsExported() && !f.IsZero() {
				matches = append(matches, name)
			}
		}
	}

	return len(matches)
}

// ReflectOneOf use reflections on struct to check if one of the fields provided contains a value
func ReflectOneOf(v interface{}, names ...string) bool {
	if !structs.IsStruct(v) {
		panic("v is not a struct")
	}

	return ReflectCountOfNotZero(v, names...) == 1
}
