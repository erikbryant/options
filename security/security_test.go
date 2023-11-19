package security

import (
	"testing"
)

func TestHasOptions(t *testing.T) {
	var answer bool
	var security Security
	var contract Contract

	answer = security.HasOptions()
	if answer != false {
		t.Errorf("ERROR: For %v expected %v, got %v", security, false, answer)
	}

	security.Puts = append(security.Puts, contract)
	answer = security.HasOptions()
	if answer != false {
		t.Errorf("ERROR: For %v expected %v, got %v", security, false, answer)
	}

	security.Calls = append(security.Calls, contract)
	answer = security.HasOptions()
	if answer != true {
		t.Errorf("ERROR: For %v expected %v, got %v", security, true, answer)
	}
}
