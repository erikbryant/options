package utils

import (
	"sort"
	"unicode"
)

// IsLetter returns true if all chars in string are alpha, false otherwise.
func IsLetter(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(char) {
			return false
		}
	}

	return true
}

// Remove returns list with any entries from skip removed
func Remove(list, skip []string) []string {
	m := make(map[string]int)

	for _, val := range list {
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
