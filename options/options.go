package options

import (
	"fmt"
	"github.com/erikbryant/options/cboe"
	"github.com/erikbryant/options/eoddata"
	"github.com/erikbryant/options/finnhub"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/options/tradeking"
	"github.com/erikbryant/options/utils"
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

// Securities accumulates stock/option data for the given tickers and returns it in a list of sec.
func Securities(tickers []string) ([]security.Security, error) {
	var securities []security.Security

	for _, ticker := range tickers {
		fmt.Printf("\r%s    ", ticker)
		sec, err := getSecurity(ticker)
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

// primary indicates which source to favor. If it starts throttling we switch to the other.
var primary = "finnhub"

// getStock retrieves stock data for the given ticker. It load balances across multiple providers.
func getStock(sec security.Security) (security.Security, error) {
	var retryable bool
	var err error

	for {
		if primary == "finnhub" {
			sec, retryable, err = finnhub.GetStock(sec)
			if err == nil || !retryable {
				break
			}
			fmt.Println("Finnhub is throttling, switching to TradeKing")
			primary = "tradeking"
		}
		sec, retryable, err = tradeking.GetStock(sec)
		if err == nil || !retryable {
			break
		}
		fmt.Println("TradeKing is throttling, switching to Finnhub")
		primary = "finnhub"
		time.Sleep(6 * time.Second)
	}

	return sec, nil
}

// getSecurity accumulates stock/option data for the given ticker and returns it in a sec.
func getSecurity(ticker string) (security.Security, error) {
	var sec security.Security

	sec.Ticker = ticker

	// Fetch data
	sec, err := tradeking.GetOptions(sec)
	if err != nil {
		return sec, fmt.Errorf("Error getting options %s %s", sec.Ticker, err)
	}
	sec, err = getStock(sec)
	if err != nil {
		return sec, fmt.Errorf("Error getting stock %s %s", sec.Ticker, err)
	}
	sec.EarningsDate = earnings[sec.Ticker]

	// Synthetic data
	for put := range sec.Puts {
		sec.Puts[put].PriceBasisDelta = sec.Price - (sec.Puts[put].Strike - sec.Puts[put].Bid)
		sec.Puts[put].LastTradeDays = int64(time.Now().Sub(sec.Puts[put].LastTradeDate).Hours() / 24)
		sec.Puts[put].BidStrikeRatio = sec.Puts[put].Bid / sec.Puts[put].Strike * 100
		sec.Puts[put].BidPriceRatio = sec.Puts[put].Bid / sec.Price * 100
		sec.Puts[put].SafetySpread = (sec.Price - (sec.Puts[put].Strike - sec.Puts[put].Bid)) / sec.Price * 100
		sec.Puts[put].CallSpread = sec.CallSpread(sec.Puts[put].Expiration)
	}
	for call := range sec.Calls {
		sec.Calls[call].PriceBasisDelta = sec.Price - (sec.Calls[call].Strike - sec.Calls[call].Bid)
		sec.Calls[call].LastTradeDays = int64(time.Now().Sub(sec.Calls[call].LastTradeDate).Hours() / 24)
		sec.Calls[call].BidStrikeRatio = sec.Calls[call].Bid / sec.Calls[call].Strike * 100
		sec.Calls[call].BidPriceRatio = sec.Calls[call].Bid / sec.Price * 100
		sec.Calls[call].SafetySpread = (sec.Price - (sec.Calls[call].Strike - sec.Calls[call].Bid)) / sec.Price * 100
		sec.Calls[call].CallSpread = sec.CallSpread(sec.Calls[call].Expiration)
	}

	return sec, nil
}

// FindSecuritiesWithOptions re-scans all known securities to see which have options and writes them to 'useFile.options.csv'.
func FindSecuritiesWithOptions(useFile string) ([]string, error) {
	securities, err := eoddata.USEquities(useFile)
	if err != nil {
		return nil, fmt.Errorf("Error loading US equity list %s", err)
	}

	securities2, err := cboe.WeeklyOptions()
	if err != nil {
		return nil, fmt.Errorf("Error loading CBOE weekly options list %s", err)
	}

	securities = utils.Combine(securities, securities2, []string{})

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

	options := []string{}
	for _, key := range securities {
		var sec security.Security
		sec.Ticker = key
		sec, err = tradeking.GetOptions(sec)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !sec.HasOptions() {
			fmt.Println("Security does not have options", sec.Ticker)
			continue
		}

		// We are looking for weekly (or more frequent) options, so the period
		// should be 7. But, if the exchange is closed on a Friday then the
		// expiration moves to Thursday. That means there are now 8 days between
		// it and the next expiration.
		period, err := sec.ExpirationPeriod()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if period > 8 {
			fmt.Println("Security expiration dates are too infrequent", sec.Ticker, period)
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
