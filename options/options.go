package options

import (
	"fmt"
	"github.com/erikbryant/options/date"
	"time"

	"github.com/erikbryant/options/finnhub"
	"github.com/erikbryant/options/marketData"
	"github.com/erikbryant/options/security"
)

// getStock returns stock data for the given security
func getStock(sec *security.Security) error {
	return finnhub.GetStock(sec)
}

// getOptions accumulates option data for the given ticker and returns it in a security
func getOptions(sec *security.Security, expiration string) error {
	// Fetch data
	err := marketData.GetOptions(sec, expiration)
	if err != nil {
		return fmt.Errorf("error getting options %s %s", sec.Ticker, err)
	}

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

	return nil
}

// Securities accumulates stock/option data for the given tickers and returns it in a list of sec
func Securities(tickers []string, expiration string, maxPrice float64) ([]security.Security, error) {
	startDate := date.Previous(time.Monday)
	endDate := date.Previous(time.Friday)
	fmt.Printf("Using startDate: %s, endDate %s for previous week's price %%change\n\n", startDate, endDate)

	var securities []security.Security

	statMaxPrice := []string{}
	statGetFail := []string{}
	statNoOptions := []string{}

	for _, ticker := range tickers {
		fmt.Printf("\r%s    ", ticker)

		sec := security.Security{
			Ticker: ticker,
		}

		err := getStock(&sec)
		if err != nil {
			return nil, fmt.Errorf("error getting stock %s %s", sec.Ticker, err)
		}

		if sec.Price >= maxPrice {
			// fmt.Printf("Skipping %s due to high price %0.2f > %0.2f\n", sec.Ticker, sec.Price, maxPrice)
			statMaxPrice = append(statMaxPrice, sec.Ticker)
			continue
		}

		sec.EarningsDate = finnhub.Earnings(sec.Ticker)

		err = getOptions(&sec, expiration)
		if err != nil {
			fmt.Printf("Error getting options: %s\n", err)
			statGetFail = append(statGetFail, sec.Ticker)
			continue
		}

		if !sec.HasOptions() {
			fmt.Printf("WARNING: %s has no options\n", sec.Ticker)
			statNoOptions = append(statNoOptions, sec.Ticker)
			continue
		}

		sec.PriceChangePct, err = marketData.PctChange(sec.Ticker, startDate, endDate)
		if err != nil {
			fmt.Println(err)
		}

		securities = append(securities, sec)
	}

	fmt.Printf("\r%d of %d tickers loaded\n\n", len(securities), len(tickers))
	fmt.Printf("  Rejected for price too high (%d): %v\n", len(statMaxPrice), statMaxPrice)
	fmt.Printf("  Rejected for get failure    (%d): %v\n", len(statGetFail), statGetFail)
	fmt.Printf("  Rejected for no options     (%d): %v\n", len(statNoOptions), statNoOptions)

	return securities, nil
}
