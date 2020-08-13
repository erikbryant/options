package security

import (
	"testing"
)

func TestOtmPutStrike(t *testing.T) {
	testCases := []struct {
		testName    string
		security    Security
		expected    float64
		shouldError bool
	}{
		{
			"OTM in strikes",
			Security{
				Ticker:  "TST",
				Price:   1.01,
				Strikes: []float64{0.5, 1.0, 1.5},
				Puts:    []Contract{},
			},
			1.0,
			false,
		},
		{
			"OTM above strikes",
			Security{
				Ticker:  "TST",
				Price:   22.34,
				Strikes: []float64{0.5, 1.0, 1.5},
				Puts:    []Contract{},
			},
			1.5,
			false,
		},
		{
			"OTM below strikes",
			Security{
				Ticker:  "TST",
				Price:   0.45,
				Strikes: []float64{0.5, 1.0, 1.5},
			},
			1.0,
			true,
		},
		{
			"ATM",
			Security{
				Ticker:  "TST",
				Price:   1.0,
				Strikes: []float64{0.5, 1.0, 1.5},
			},
			1.0,
			false,
		},
		{
			"no strikes",
			Security{
				Ticker: "TST",
				Price:  0.45,
			},
			1.0,
			true,
		},
		{
			"empty security",
			Security{},
			0,
			true,
		},
	}

	for _, testCase := range testCases {
		answer, err := testCase.security.otmPutStrike()
		if testCase.shouldError {
			if err == nil {
				t.Errorf("For '%s' got unexpected error %s", testCase.testName, err)
			}
			continue
		}

		if err != nil {
			t.Errorf("For '%s' got unexpected error %s", testCase.testName, err)
			continue
		}

		if answer != testCase.expected {
			t.Errorf("For '%s' expected %f got %f", testCase.testName, testCase.expected, answer)
		}
	}
}

func TestItmPutStrike(t *testing.T) {
	testCases := []struct {
		testName    string
		security    Security
		expected    float64
		shouldError bool
	}{
		{
			"OTM in strikes",
			Security{
				Ticker:  "TST",
				Price:   1.01,
				Strikes: []float64{0.5, 1.0, 1.5},
				Puts:    []Contract{},
			},
			1.5,
			false,
		},
		{
			"OTM above strikes",
			Security{
				Ticker:  "TST",
				Price:   22.34,
				Strikes: []float64{0.5, 1.0, 1.5},
				Puts:    []Contract{},
			},
			0,
			true,
		},
		{
			"OTM below strikes",
			Security{
				Ticker:  "TST",
				Price:   0.45,
				Strikes: []float64{0.5, 1.0, 1.5},
			},
			0.5,
			false,
		},
		{
			"ATM",
			Security{
				Ticker:  "TST",
				Price:   1.0,
				Strikes: []float64{0.5, 1.0, 1.5},
			},
			1.5,
			false,
		},
		{
			"no strikes",
			Security{
				Ticker: "TST",
				Price:  0.45,
			},
			0,
			true,
		},
		{
			"empty security",
			Security{},
			0,
			true,
		},
	}

	for _, testCase := range testCases {
		answer, err := testCase.security.itmPutStrike()
		if testCase.shouldError {
			if err == nil {
				t.Errorf("For '%s' got unexpected error %s", testCase.testName, err)
			}
			continue
		}

		if err != nil {
			t.Errorf("For '%s' got unexpected error %s", testCase.testName, err)
			continue
		}

		if answer != testCase.expected {
			t.Errorf("For '%s' expected %f got %f", testCase.testName, testCase.expected, answer)
		}
	}
}
