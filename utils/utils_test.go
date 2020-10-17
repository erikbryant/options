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

func TestCombine(t *testing.T) {
	testCases := []struct {
		list1    []string
		list2    []string
		skip     []string
		expected []string
	}{
		// Blank and a skip.
		{
			[]string{"", "A"},
			[]string{"B", "C"},
			[]string{"C"},
			[]string{"A", "B"},
		},
		// All empty.
		{
			[]string{},
			[]string{},
			[]string{},
			[]string{},
		},
		// list1 empty.
		{
			[]string{},
			[]string{"A", "B"},
			[]string{},
			[]string{"A", "B"},
		},
		// Unsorted input.
		{
			[]string{"Z", "M", "C"},
			[]string{"B", "A"},
			[]string{},
			[]string{"A", "B", "C", "M", "Z"},
		},
		// Duplicates.
		{
			[]string{"Z", "Z", "C"},
			[]string{"B", "C"},
			[]string{},
			[]string{"B", "C", "Z"},
		},
		// Skip deletes all.
		{
			[]string{"Z", "Z", "C"},
			[]string{"B", "C"},
			[]string{"B", "C", "Z"},
			[]string{},
		},
		// Skip has values not in either list.
		{
			[]string{"Z", "Z", "C"},
			[]string{"B", "C"},
			[]string{"W", "Y", "X"},
			[]string{"B", "C", "Z"},
		},
	}

	for _, testCase := range testCases {
		answer := Combine(testCase.list1, testCase.list2, testCase.skip)
		if !equal(answer, testCase.expected) {
			t.Errorf("For %v, %v, %v expected %v, got %v", testCase.list1, testCase.list2, testCase.skip, testCase.expected, answer)
		}
	}
}
