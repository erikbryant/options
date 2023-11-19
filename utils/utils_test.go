package utils

import (
	"testing"
)

// equal returns true if the two lists are identical, false otherwise.
func equal(list1, list2 []string) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i := range list1 {
		if list1[i] != list2[i] {
			return false
		}
	}

	return true
}

func TestRemove(t *testing.T) {
	testCases := []struct {
		list     []string
		skip     []string
		expected []string
	}{
		// Blank and a skip.
		{
			[]string{"", "A"},
			[]string{"C"},
			[]string{"A"},
		},
		// All empty.
		{
			[]string{},
			[]string{},
			[]string{},
		},
		// Unsorted input.
		{
			[]string{"Z", "M", "C"},
			[]string{},
			[]string{"C", "M", "Z"},
		},
		// Duplicates.
		{
			[]string{"Z", "Z", "C"},
			[]string{},
			[]string{"C", "Z"},
		},
		// Skip deletes all.
		{
			[]string{"Z", "Z", "C"},
			[]string{"B", "C", "Z"},
			[]string{},
		},
		// Skip has values not in either list.
		{
			[]string{"Z", "Z", "C"},
			[]string{"W", "Y", "X"},
			[]string{"C", "Z"},
		},
	}

	for _, testCase := range testCases {
		answer := Remove(testCase.list, testCase.skip)
		if !equal(answer, testCase.expected) {
			t.Errorf("For %v, %v expected %v, got %v", testCase.list, testCase.skip, testCase.expected, answer)
		}
	}
}
