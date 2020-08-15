package yahoo

import (
	"testing"
)

// This is essentially identical to TestGetFmtString & TestGet
func TestGetRawFloat(t *testing.T) {
	testCases := []struct {
		json        map[string]interface{}
		key         string
		expected    float64
		shouldError bool
	}{
		// Success
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
		// Missing the top-level key
		{
			map[string]interface{}{
				"foob": map[string]interface{}{
					"raw": 3.2,
				},
			},
			"foo2",
			0,
			false,
		},
		// Missing the second-level key
		{
			map[string]interface{}{
				"foo3": map[string]interface{}{
					"rare": 3.2,
				},
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
			false,
		},
		// Nil map
		{
			map[string]interface{}{},
			"foo5",
			0,
			false,
		},
		// Nil second-level map
		{
			map[string]interface{}{
				"foo6": nil,
			},
			"foo6",
			0,
			false,
		},
		// Wrong type for float
		{
			map[string]interface{}{
				"foo7": map[string]interface{}{
					"raw": nil,
				},
			},
			"foo7",
			3.2,
			true,
		},
	}

	for _, testCase := range testCases {
		answer, err := getRawFloat(testCase.json, testCase.key)
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

// This is essentially identical to TestGetRawFloat & TestGet
func TestGetFmtString(t *testing.T) {
	testCases := []struct {
		json        map[string]interface{}
		key         string
		expected    string
		shouldError bool
	}{
		// Success
		{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"fmt": "bar",
				},
			},
			"foo",
			"bar",
			false,
		},
		// Missing the top-level key
		{
			map[string]interface{}{
				"foob": map[string]interface{}{
					"fmt": "bar",
				},
			},
			"foo",
			"bar",
			true,
		},
		// Missing the second-level key
		{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"fmts": "bar",
				},
			},
			"foo",
			"bar",
			true,
		},
		// Nil
		{
			nil,
			"foo",
			"bar",
			true,
		},
		// Nil map
		{
			map[string]interface{}{},
			"foo",
			"bar",
			true,
		},
		// Nil second-level map
		{
			map[string]interface{}{
				"foo": nil,
			},
			"foo",
			"bar",
			true,
		},
		// Wrong type for float
		{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"fmt": nil,
				},
			},
			"foo",
			"bar",
			true,
		},
	}

	for _, testCase := range testCases {
		answer, err := getFmtString(testCase.json, testCase.key)
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
			t.Errorf("For %s expected %s got %s", testCase.key, testCase.expected, answer)
		}
	}
}

// This is essentially identical to TestGetRawFloat & TestGetFmtString
// func TestGet(t *testing.T) {
// 	testCases := []struct {
// 		json        map[string]interface{}
// 		key         string
// 		expected    interface{}
// 		shouldError bool
// 	}{
// 		// Success
// 		{
// 			map[string]interface{}{
// 				"foo": map[string]interface{}{
// 					"raw": 3.2,
// 				},
// 			},
// 			"foo",
// 			map[string]float64{
// 				"raw": 3.2,
// 			},
// 			false,
// 		},
// 		// Missing the top-level key
// 		{
// 			map[string]interface{}{
// 				"foob": map[string]interface{}{
// 					"raw": 3.2,
// 				},
// 			},
// 			"foo",
// 			nil,
// 			false,
// 		},
// 		// Missing the second-level key
// 		{
// 			map[string]interface{}{
// 				"foo": map[string]interface{}{
// 					"rare": 3.2,
// 				},
// 			},
// 			"foo",
// 			3.2,
// 			true,
// 		},
// 		// Nil
// 		{
// 			nil,
// 			"foo",
// 			3.2,
// 			true,
// 		},
// 		// Nil map
// 		{
// 			map[string]interface{}{},
// 			"foo",
// 			3.2,
// 			true,
// 		},
// 		// Nil second-level map
// 		{
// 			map[string]interface{}{
// 				"foo": nil,
// 			},
// 			"foo",
// 			3.2,
// 			true,
// 		},
// 		// Wrong type for float
// 		{
// 			map[string]interface{}{
// 				"foo": map[string]interface{}{
// 					"raw": nil,
// 				},
// 			},
// 			"foo",
// 			3.2,
// 			true,
// 		},
// 	}
//
// 	for _, testCase := range testCases {
// 		answer, err := get(testCase.json, testCase.key)
// 		if testCase.shouldError {
// 			if err == nil {
// 				t.Errorf("For %s it should have errored but did not", testCase.key)
// 			}
// 			continue
// 		}
// 		if err != nil {
// 			t.Errorf("For %s got unexpected error %s", testCase.key, err)
// 			continue
// 		}
// 		if answer != testCase.expected {
// 			t.Errorf("For %s expected %f got %f", testCase.key, testCase.expected, answer)
// 		}
// 	}
// }

func TestParseContracts(t *testing.T) {

}

func TestParsePrice(t *testing.T) {

}

func TestParseStrikes(t *testing.T) {

}

func TestExtractJSON(t *testing.T) {
	testCases := []struct {
		text        string
		shouldError bool
	}{
		// Success
		{
			headerToken + "{\"key\": \"val\"}" + ";\n}(this));",
			false,
		},
		// No header token
		{
			"foo",
			true,
		},
		// No footer token
		{
			headerToken + " foo",
			true,
		},
		// Not valid JSON
		{
			headerToken + "{45}" + ";\n}(this));",
			true,
		},
	}

	for _, testCase := range testCases {
		answer, err := extractJSON(testCase.text)
		if testCase.shouldError {
			if err == nil {
				t.Errorf("For %s it should have errored but did not", testCase.text)
			}
			continue
		}

		if err != nil {
			t.Errorf("For %s got unexpected error %s", testCase.text, err)
			continue
		}

		if answer["key"] != "val" {
			t.Errorf("For %s expected val got %s", testCase.text, answer)
		}
	}
}

func TestExtractOCS(t *testing.T) {
	testCases := []struct {
		json        map[string]interface{}
		key         string
		expected    string
		shouldError bool
	}{
		// Success
		{
			map[string]interface{}{
				"context": map[string]interface{}{
					"dispatcher": map[string]interface{}{
						"stores": map[string]interface{}{
							"OptionContractsStore": map[string]interface{}{
								"foo": "bar",
							},
						},
					},
				},
			},
			"foo",
			"bar",
			false,
		},
		// Missing OptionContractsStore
		{
			map[string]interface{}{
				"context": map[string]interface{}{
					"dispatcher": map[string]interface{}{
						"stores": map[string]interface{}{
							"FooStore": map[string]interface{}{
								"foo": "bar",
							},
						},
					},
				},
			},
			"foo",
			"bar",
			true,
		},
		// Missing stores
		{
			map[string]interface{}{
				"context": map[string]interface{}{
					"dispatcher": map[string]interface{}{
						"xyzstores": map[string]interface{}{},
					},
				},
			},
			"foo",
			"bar",
			true,
		},
		// Missing dispatcher
		{
			map[string]interface{}{
				"context": map[string]interface{}{
					"xyzdispatcher": map[string]interface{}{},
				},
			},
			"foo",
			"bar",
			true,
		},
		// Missing context
		{
			map[string]interface{}{},
			"foo",
			"bar",
			true,
		},
		// Nil
		{
			nil,
			"foo",
			"bar",
			true,
		},
	}

	for _, testCase := range testCases {
		answer, err := extractOCS(testCase.json)
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

		if answer[testCase.key] != testCase.expected {
			t.Errorf("For %v expected %s got %s", testCase.json, testCase.expected, answer)
		}
	}
}
