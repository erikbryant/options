package options

import (
	"fmt"
	"github.com/erikbryant/options/cboe"
	"github.com/erikbryant/options/eoddata"
	"github.com/erikbryant/options/finnhub"
	sec "github.com/erikbryant/options/security"
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

// Securities accumulates stock/option data for the given tickers and returns it in a list of Security.
func Securities(tickers []string) ([]sec.Security, error) {
	var securities []sec.Security

	for _, ticker := range tickers {
		fmt.Printf("\r%s    ", ticker)
		security, err := Security(ticker)
		if err != nil {
			fmt.Printf("Error getting security data %s\n", err)
			continue
		}

		if !security.HasOptions() {
			continue
		}

		securities = append(securities, security)
	}

	fmt.Printf("\r%d of %d tickers loaded\n\n", len(securities), len(tickers))

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
		time.Sleep(6 * time.Second)
	}

	return security, nil
}

// Security accumulates stock/option data for the given ticker and returns it in a Security.
func Security(ticker string) (sec.Security, error) {
	var security sec.Security

	security.Ticker = ticker

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
		security.Puts[put].BidPriceRatio = security.Puts[put].Bid / security.Price * 100
		security.Puts[put].SafetySpread = (security.Price - (security.Puts[put].Strike - security.Puts[put].Bid)) / security.Price * 100
		security.Puts[put].CallSpread = security.CallSpread(security.Puts[put].Expiration)
	}
	for call := range security.Calls {
		security.Calls[call].PriceBasisDelta = security.Price - (security.Calls[call].Strike - security.Calls[call].Bid)
		security.Calls[call].LastTradeDays = int64(time.Now().Sub(security.Calls[call].LastTradeDate).Hours() / 24)
		security.Calls[call].BidStrikeRatio = security.Calls[call].Bid / security.Calls[call].Strike * 100
		security.Calls[call].BidPriceRatio = security.Calls[call].Bid / security.Price * 100
		security.Calls[call].SafetySpread = (security.Price - (security.Calls[call].Strike - security.Calls[call].Bid)) / security.Price * 100
		security.Calls[call].CallSpread = security.CallSpread(security.Calls[call].Expiration)
	}

	return security, nil
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
		var security sec.Security
		security.Ticker = key
		security, err = tradeking.GetOptions(security)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !security.HasOptions() {
			fmt.Println("Security does not have options", security.Ticker)
			continue
		}

		// We are looking for weekly (or more frequent) options, so the period
		// should be 7. But, if the exchange is closed on a Friday then the
		// expiration moves to Thursday. That means there are now 8 days between
		// it and the next expiration.
		period, err := security.ExpirationPeriod()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if period > 8 {
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
