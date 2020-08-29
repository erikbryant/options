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

// Securities accumulates stock/option data for the given tickers and returns it in a list of Security.
func Securities(tickers []string) ([]sec.Security, error) {
	var securities []sec.Security

	for _, ticker := range tickers {
		security, err := Security(ticker)
		if err != nil {
			fmt.Println("Error getting security data", err)
			continue
		}

		if !security.HasOptions() {
			continue
		}

		securities = append(securities, security)
	}

	fmt.Printf("%d of %d tickers loaded\n\n", len(securities), len(tickers))

	return securities, nil
}

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

// FindSecuritiesWithOptions re-scans all known securities to see which have options.
func FindSecuritiesWithOptions(useFile, optionsFile string) ([]string, error) {
	securities, err := eoddata.USEquities(useFile)
	if err != nil {
		return nil, fmt.Errorf("Error loading US equity list %s", err)
	}

	var keys []string
	for key := range securities {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	outFile := strings.Replace(useFile, ".csv", ".options.csv", 1)
	f, err := os.Create(outFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.WriteString("Symbol\n")
	if err != nil {
		return nil, err
	}

	knownOptions, err := sec.GetFile(optionsFile)
	if err != nil {
		return nil, err
	}

	optionable := make(map[string]string)
	for _, o := range knownOptions {
		optionable[o] = o
	}

	options := []string{}
	for _, key := range keys {
		// If we already know it has options, skip it
		if optionable[key] != "" {
			continue
		}

		security, err := Security(key)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if !security.HasOptions() {
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
