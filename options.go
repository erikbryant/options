package options

import (
	"fmt"
	"github.com/erikbryant/options/eoddata"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/options/yahoo"
	"os"
	"sort"
	"strings"
)

// Security accumulates stock/option data for the given ticker and returns it in a Security.
func Security(ticker string) (sec.Security, error) {
	var security sec.Security

	security.Ticker = ticker

	security, err := yahoo.Symbol(security)
	if err != nil {
		return security, fmt.Errorf("Error getting security %s %s", security.Ticker, err)
	}

	return security, nil
}

// SecuritiesWithOptions loads the cached list of all securities known to have options.
func SecuritiesWithOptions(optionsFile string) ([]string, error) {
	secs, err := sec.GetFile(optionsFile)
	if err != nil {
		return nil, err
	}

	// Strip the column header row.
	return secs[1:], nil
}

// FindSecuritiesWithOptions re-scans all known securities to see which have options.
func FindSecuritiesWithOptions(useFile string) ([]string, error) {
	securities, err := eoddata.USEquities(useFile)
	if err != nil {
		return nil, fmt.Errorf("Error loading US equity list %s", err)
	}

	var keys []string
	for key := range securities {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	optionsFile := strings.Replace(useFile, ".csv", ".options.csv", 1)
	f, err := os.Create(optionsFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.WriteString("Symbol\n")
	if err != nil {
		return nil, err
	}

	options := []string{}
	for _, key := range keys {
		hasOptions, err := yahoo.SymbolHasOptions(key)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if !hasOptions {
			continue
		}

		options = append(options, key)
		fmt.Println(key)

		_, err = f.WriteString(key + "\n")
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}
