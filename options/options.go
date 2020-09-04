package options

import (
	"fmt"
	"github.com/erikbryant/options/csv"
	"github.com/erikbryant/options/eoddata"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/options/tradeking"
	// "github.com/erikbryant/options/yahoo"
	"github.com/erikbryant/options/finnhub"
	"os"
	"strings"
	"time"
)

var (
	earnings map[string]string
)

// Init initializes the internal state of package options.
func Init(start, end string) (err error) {
	earnings, err = finnhub.EarningDates(start, end)
	return
}

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

// primary indicates which source to favor. If it starts throttling we switch to the other.
var primary = "finnhub"

// getStock retrieves stock data for the given ticker. It load balances across multiple providers.
func getStock(security sec.Security) (sec.Security, error) {
	var retryable bool
	var err error

	for {
		if primary == "finnhub" {
			security, retryable, err = finnhub.GetStock(security)
			if err == nil || !retryable {
				break
			}
			fmt.Println("Finnhub is throttling, switching to TradeKing")
			primary = "tradeking"
		}
		security, retryable, err = tradeking.GetStock(security)
		if err == nil || !retryable {
			break
		}
		fmt.Println("TradeKing is throttling, switching to Finnhub")
		primary = "finnhub"
	}

	return security, nil
}

// Security accumulates stock/option data for the given ticker and returns it in a Security.
func Security(ticker string) (sec.Security, error) {
	var security sec.Security

	security.Ticker = ticker
	fmt.Println(security.Ticker)

	// Fetch data
	security, err := tradeking.GetOptions(security)
	if err != nil {
		return security, fmt.Errorf("Error getting options %s %s", security.Ticker, err)
	}
	security, err = getStock(security)
	if err != nil {
		return security, fmt.Errorf("Error getting stock %s %s", security.Ticker, err)
	}
	security.EarningsDate = earnings[security.Ticker]

	// Synthetic data
	for put := range security.Puts {
		security.Puts[put].PriceBasisDelta = security.Price - (security.Puts[put].Strike - security.Puts[put].Bid)
		security.Puts[put].LastTradeDays = int64(time.Now().Sub(security.Puts[put].LastTradeDate).Hours() / 24)
		security.Puts[put].BidStrikeRatio = security.Puts[put].Bid / security.Puts[put].Strike * 100
		security.Puts[put].SafetySpread = (security.Price - (security.Puts[put].Strike - security.Puts[put].Bid)) / security.Price * 100
	}

	return security, nil
}

// FindSecuritiesWithOptions re-scans all known securities to see which have options.
func FindSecuritiesWithOptions(useFile, optionsFile string) ([]string, error) {
	securities, err := eoddata.USEquities(useFile)
	if err != nil {
		return nil, fmt.Errorf("Error loading US equity list %s", err)
	}

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

	optionable := make(map[string]string)
	if optionsFile != "" {
		knownOptions, err := csv.GetFile(optionsFile)
		if err != nil {
			return nil, err
		}
		for _, o := range knownOptions {
			optionable[o] = o
		}
	}

	options := []string{}
	for _, key := range securities {
		// If we already know it has options, skip it
		if optionable[key] != "" {
			continue
		}

		var security sec.Security
		security.Ticker = key
		security, err = tradeking.GetOptions(security)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !security.HasOptions() {
			continue
		}

		period, err := security.ExpirationPeriod()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if period > 7 {
			fmt.Println("Security expiration dates are too infrequent", security.Ticker, period)
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
