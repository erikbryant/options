package yahoo

import (
	"testing"
)

func TestGetRawFloat(t *testing.T) {
	testCases := []struct {
		json        map[string]interface{}
		key         string
		expected    float64
		shouldError bool
	}{
		{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"raw": 3.2,
				},
			},
			"foo",
			3.2,
			false,
		},
	}

	for _, testCase := range testCases {
		answer, err := getRawFloat(testCase.json, testCase.key)
		if err != nil && !testCase.shouldError {
			t.Errorf("For %s got unexpected error %s", testCase.key, err)
		}
		if err == nil && testCase.shouldError {
			t.Errorf("For %s it should have errored but did not", testCase.key)
		}
		if answer != testCase.expected {
			t.Errorf("For %s expected %f got %f", testCase.key, testCase.expected, answer)
		}
	}
}
