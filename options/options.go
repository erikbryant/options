package options

import (
	"fmt"
	"time"

	"github.com/erikbryant/options/finnhub"
	"github.com/erikbryant/options/marketData"
	"github.com/erikbryant/options/security"
)

// getStock returns stock data for the given security
func getStock(sec security.Security) (security.Security, error) {
	return finnhub.GetStock(sec)
}

// getSecurity accumulates stock/option data for the given ticker and returns it in a security
func getSecurity(ticker, expiration string) (security.Security, error) {
	var sec security.Security

	sec.Ticker = ticker

	// Fetch data
	sec, err := marketData.GetOptions(sec, expiration)
	if err != nil {
		return sec, fmt.Errorf("error getting options %s %s", sec.Ticker, err)
	}
	sec, err = getStock(sec)
	if err != nil {
		return sec, fmt.Errorf("error getting stock %s %s", sec.Ticker, err)
	}
	sec.EarningsDate = finnhub.Earnings(sec.Ticker)

	// Synthetic data. Use the index to access the option instead of having
	// range return the option, since range returns a COPY of the option.
	for put := range sec.Puts {
		sec.Puts[put].PriceBasisDelta = sec.Price - (sec.Puts[put].Strike - sec.Puts[put].Bid)
		sec.Puts[put].LastTradeDays = int64(time.Since(sec.Puts[put].LastTradeDate).Hours() / 24)
		sec.Puts[put].BidStrikeRatio = sec.Puts[put].Bid / sec.Puts[put].Strike * 100
		sec.Puts[put].BidPriceRatio = sec.Puts[put].Bid / sec.Price * 100
		sec.Puts[put].SafetySpread = (sec.Price - (sec.Puts[put].Strike - sec.Puts[put].Bid)) / sec.Price * 100
		sec.Puts[put].CallSpread = sec.CallSpread(sec.Puts[put].Expiration)
	}
	for call := range sec.Calls {
		sec.Calls[call].PriceBasisDelta = sec.Price - (sec.Calls[call].Strike - sec.Calls[call].Bid)
		sec.Calls[call].LastTradeDays = int64(time.Since(sec.Calls[call].LastTradeDate).Hours() / 24)
		sec.Calls[call].BidStrikeRatio = sec.Calls[call].Bid / sec.Calls[call].Strike * 100
		sec.Calls[call].BidPriceRatio = sec.Calls[call].Bid / sec.Price * 100
		sec.Calls[call].SafetySpread = (sec.Price - (sec.Calls[call].Strike - sec.Calls[call].Bid)) / sec.Price * 100
		sec.Calls[call].CallSpread = sec.CallSpread(sec.Calls[call].Expiration)
	}

	return sec, nil
}

// Securities accumulates stock/option data for the given tickers and returns it in a list of sec
func Securities(tickers []string, expiration string) ([]security.Security, error) {
	var securities []security.Security

	for _, ticker := range tickers {
		fmt.Printf("\r%s    ", ticker)
		sec, err := getSecurity(ticker, expiration)
		if err != nil {
			fmt.Printf("Error getting security data %s\n", err)
			continue
		}

		if !sec.HasOptions() {
			continue
		}

		securities = append(securities, sec)
	}

	fmt.Printf("\r%d of %d tickers loaded\n\n", len(securities), len(tickers))

	return securities, nil
}
