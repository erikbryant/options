package utils

import (
	"sort"
)

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

	// We pulled these out of a map, so they are now unordered
	sort.Strings(result)

	return result
}
