package options

import (
	"fmt"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/options/yahoo"
)

// GetSecurity gets all of the relevant data into the security.
func GetSecurity(ticker string) (security.Security, error) {
	var security security.Security

	security.Ticker = ticker

	optionContractsStore, err := yahoo.Symbol(ticker)
	if err != nil {
		return security, fmt.Errorf("Error getting security %s %s", ticker, err)
	}

	err = security.ParseOCS(optionContractsStore)
	if err != nil {
		return security, fmt.Errorf("Error parsing OCS %s", err)
	}

	return security, nil
}
