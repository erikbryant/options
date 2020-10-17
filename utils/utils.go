package utils

import (
	"sort"
	"unicode"
)

// AlphaNumeric returns true if all chars in string are alphanumeric, false otherwise.
func AlphaNumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(char) {
			return false
		}
	}

	return true
}

// Combine merges two lists into one, removes any elements that are in skip, and returns the sorted remainder.
func Combine(list1, list2 []string, skip []string) []string {
	m := make(map[string]int)

	for _, val := range list1 {
		m[val] = 1
	}

	for _, val := range list2 {
		m[val] = 1
	}

	for _, val := range skip {
		delete(m, val)
	}

	var result []string

	for key := range m {
		if key == "" {
			continue
		}
		result = append(result, key)
	}

	sort.Strings(result)

	return result
}
