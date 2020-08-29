package yahoo

import (
	"testing"
)

func TestGetFloat(t *testing.T) {
	testCases := []struct {
		json        map[string]interface{}
		key         string
		expected    float64
		shouldError bool
	}{
		// Success
		{
			map[string]interface{}{
				"foo": 3.2,
			},
			"foo",
			3.2,
			false,
		},
		// Missing the top-level key
		{
			map[string]interface{}{
				"foob": 3.2,
			},
			"foo2",
			0,
			true,
		},
		// Missing the second-level key
		{
			map[string]interface{}{
				"foo3": nil,
			},
			"foo3",
			3.2,
			true,
		},
		// Nil
		{
			nil,
			"foo4",
			0,
			true,
		},
		// Nil map
		{
			map[string]interface{}{},
			"foo5",
			0,
			true,
		},
		// Nil second-level map
		{
			map[string]interface{}{
				"foo6": nil,
			},
			"foo6",
			0,
			true,
		},
		// Wrong type for float
		{
			map[string]interface{}{
				"foo7": "not a float",
			},
			"foo7",
			3.2,
			true,
		},
	}

	for _, testCase := range testCases {
		answer, err := getFloat(testCase.json, testCase.key)
		if testCase.shouldError {
			if err == nil {
				t.Errorf("For %s it should have errored but did not", testCase.key)
			}
			continue
		}
		if err != nil {
			t.Errorf("For %s got unexpected error %s", testCase.key, err)
			continue
		}
		if answer != testCase.expected {
			t.Errorf("For %s expected %f got %f", testCase.key, testCase.expected, answer)
		}
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		json        map[string]interface{}
		key         string
		expected    interface{}
		shouldError bool
	}{
		// Success
		// {
		// 	map[string]interface{}{
		// 		"foo": map[string]interface{}{
		// 			"raw": 3.2,
		// 		},
		// 	},
		// 	"foo",
		// 	map[string]interface{}{
		// 		"raw": 3.2,
		// 	},
		// 	false,
		// },
		// Missing the top-level key
		{
			map[string]interface{}{
				"foob": map[string]interface{}{
					"raw": 3.2,
				},
			},
			"foo2",
			nil,
			true,
		},
		// Success
		{
			map[string]interface{}{
				"foo3": "val",
			},
			"foo3",
			"val",
			false,
		},
		// Nil
		{
			nil,
			"foo4",
			3.2,
			true,
		},
		// Nil map
		{
			map[string]interface{}{},
			"foo5",
			3.2,
			true,
		},
		// Nil second-level map
		{
			map[string]interface{}{
				"foo6": nil,
			},
			"foo6",
			nil,
			false,
		},
	}

	for _, testCase := range testCases {
		answer, err := get(testCase.json, testCase.key)
		if testCase.shouldError {
			if err == nil {
				t.Errorf("For %s it should have errored but did not", testCase.key)
			}
			continue
		}
		if err != nil {
			t.Errorf("For %s got unexpected error %s", testCase.key, err)
			continue
		}
		if answer != testCase.expected {
			t.Errorf("For %s expected %f got %f", testCase.key, testCase.expected, answer)
		}
	}
}

func TestParseContracts(t *testing.T) {

}

func TestParsePrice(t *testing.T) {

}

func TestParseStrikes(t *testing.T) {

}
