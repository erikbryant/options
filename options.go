package options

import (
	"fmt"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/options/yahoo"
)

// GetSecurity accumulates stock/option data for the given ticker and returns it in a Security.
func GetSecurity(ticker string) (sec.Security, error) {
	var security sec.Security

	security.Ticker = ticker

	security, err := yahoo.Symbol(security)
	if err != nil {
		return security, fmt.Errorf("Error getting security %s %s", security.Ticker, err)
	}

	return security, nil
}
