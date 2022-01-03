package cache

import (
	"testing"
)

func TestSanitize(t *testing.T) {
	testCases := []struct {
		s        string
		expected string
	}{
		{"", ""},
		{"asdf", "asdf"},
		{"/", "-"},
		{"//", "--"},
		{"a/b", "a-b"},
		{"-", "-"},
		{"/-/", "---"},
	}

	for _, testCase := range testCases {
		answer := sanitize(testCase.s)
		if answer != testCase.expected {
			t.Errorf("For %s expected %s, got %s", testCase.s, testCase.expected, answer)
		}
	}
}
