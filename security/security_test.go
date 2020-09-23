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
	if answer != false {
		t.Errorf("ERROR: For %v expected %v, got %v", security, false, answer)
	}

	security.Strikes = append(security.Strikes, 10.0)
	answer = security.HasOptions()
	if answer != true {
		t.Errorf("ERROR: For %v expected %v, got %v", security, true, answer)
	}
}

func TestExpirationPeriod(t *testing.T) {
	var security Security
	var contract Contract
	var answer int
	var err error

	// Daily expiration period.
	security.Puts = []Contract{}
	contract.Expiration = "20200925"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20200926"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20200927"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20200928"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20200929"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20200930"
	security.Puts = append(security.Puts, contract)
	answer, err = security.ExpirationPeriod()
	if err != nil {
		t.Errorf("ERROR: For %v got unexpected error %v", security, err)
	}
	if answer != 1 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, 1, answer)
	}

	// 3 day expiration period.
	security.Puts = []Contract{}
	contract.Expiration = "20200925"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20200928"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201001"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201004"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201007"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201010"
	security.Puts = append(security.Puts, contract)
	answer, err = security.ExpirationPeriod()
	if err != nil {
		t.Errorf("ERROR: For %v got unexpected error %v", security, err)
	}
	if answer != 3 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, 3, answer)
	}

	// Weekly expiration period.
	security.Puts = []Contract{}
	contract.Expiration = "20200925"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201002"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201009"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201016"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201023"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201030"
	security.Puts = append(security.Puts, contract)
	answer, err = security.ExpirationPeriod()
	if err != nil {
		t.Errorf("ERROR: For %v got unexpected error %v", security, err)
	}
	if answer != 7 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, 7, answer)
	}

	// Monthly expiration period.
	security.Puts = []Contract{}
	contract.Expiration = "20200925"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201025"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201125"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201225"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20210125"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20210225"
	security.Puts = append(security.Puts, contract)
	answer, err = security.ExpirationPeriod()
	if err != nil {
		t.Errorf("ERROR: For %v got unexpected error %v", security, err)
	}
	if answer != 31 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, 31, answer)
	}

	// Weekly expiration period, but one missing so should return 14.
	security.Puts = []Contract{}
	contract.Expiration = "20200925"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201002"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201009"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201016"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "20201030"
	security.Puts = append(security.Puts, contract)
	answer, err = security.ExpirationPeriod()
	if err != nil {
		t.Errorf("ERROR: For %v got unexpected error %v", security, err)
	}
	if answer != 14 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, 14, answer)
	}

	// Error too few expirations.
	security.Puts = []Contract{}
	answer, err = security.ExpirationPeriod()
	if err == nil {
		t.Errorf("ERROR: For %v expected error, but did not get it", security)
	}
	if answer != -1 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, -1, answer)
	}

	// Error too few expirations.
	security.Puts = []Contract{}
	contract.Expiration = "20200925"
	security.Puts = append(security.Puts, contract)
	answer, err = security.ExpirationPeriod()
	if err == nil {
		t.Errorf("ERROR: For %v expected error, but did not get it", security)
	}
	if answer != -1 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, -1, answer)
	}

	// Error, expiration periods not parsable.
	security.Puts = []Contract{}
	contract.Expiration = "foo1"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "foo2"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "foo3"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "foo4"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "foo5"
	security.Puts = append(security.Puts, contract)
	contract.Expiration = "foo6"
	security.Puts = append(security.Puts, contract)
	answer, err = security.ExpirationPeriod()
	if err == nil {
		t.Errorf("ERROR: For %v expected error, but did not get it", security)
	}
	if answer != -1 {
		t.Errorf("ERROR: For %v expected %v, got %v", security, -1, answer)
	}
}
