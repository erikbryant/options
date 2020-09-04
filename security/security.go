package security

import (
	"fmt"
	"sort"
	"time"
)

// Contract holds option data for a single expiration date.
type Contract struct {
	// Received values
	Strike        float64
	Last          float64
	Bid           float64
	Ask           float64
	Expiration    string
	LastTradeDate time.Time
	Size          int
	OpenInterest  int64
	// Derived values
	PriceBasisDelta float64 // Share price minus cost basis
	LastTradeDays   int64   // Age of last trade in days
	BidStrikeRatio  float64 // bid / strike
	SafetySpread    float64 // distance between share price and cost basis
}

// DayRange represents a single (historical) trading day.
type DayRange struct {
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Security holds data about a security and its options.
type Security struct {
	Ticker       string
	Close        DayRange
	Price        float64
	Strikes      []float64
	Puts         []Contract
	Calls        []Contract
	EarningsDate string
}

// HasOptions returns whether the security has options.
func (security *Security) HasOptions() bool {
	return len(security.Puts) != 0 && len(security.Calls) != 0 && len(security.Strikes) != 0
}

// ExpirationPeriod tries to determine the time between expiration dates.
func (security *Security) ExpirationPeriod() (int, error) {
	uniques := make(map[string]int)

	for _, put := range security.Puts {
		uniques[put.Expiration] = 1
	}

	var expirations []string

	for expiration := range uniques {
		expirations = append(expirations, expiration)
	}

	sort.Strings(expirations)

	// Look for there to be a lot of expiration dates. If there are only a few
	// then we cannot get an accurate period.
	// The choice of which dates to look at is arbitrary, but as long as we are
	// looking out at least a month we are sure to be fine.
	// But, don't look out too far. We might get into LEAPS or something with
	// very far expirations.

	const minExpirations = 5

	if len(expirations) < minExpirations {
		return -1, fmt.Errorf("Not enough expirations to determine period %s %v", security.Ticker, expirations)
	}

	var max = 0
	var prev time.Time
	var next time.Time
	var err error

	prev, err = time.Parse("20060102", expirations[0])
	if err != nil {
		return -1, fmt.Errorf("Could not parse first expiration date %s %s", err, expirations[0])
	}

	for i := 1; i < minExpirations; i++ {
		next, err = time.Parse("20060102", expirations[i])
		if err != nil {
			return -1, fmt.Errorf("Could not parse next expiration date %s %s", err, expirations[i])
		}

		days := int(next.Sub(prev).Hours() / 24)
		if days > max {
			max = days
		}

		prev = next
	}

	return max, nil
}

// PrintPut prints the put data for a single ticker in CSV or tabulated and with or without a header.
func (security *Security) PrintPut(put int, csv, header bool, expiration string) {
	var separator string

	if csv {
		separator = ", "
	} else {
		separator = "  "
	}

	if header {
		fmt.Printf("%8s", "Ticker")
		fmt.Printf(separator)
		fmt.Printf("%10s", "Expiration")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Price")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Strike")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Last")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Bid")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Ask")
		fmt.Printf(separator)
		fmt.Printf("%8s", "B/S ratio")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Safety")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Age")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Earnings")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Price+")
		fmt.Printf(separator)
		fmt.Printf("%8s", "Odd Lot")
		fmt.Printf("\n")
	}

	// TODO: If it is the weekend, then make the threshold for
	// printing higher.
	var lastTrade string
	if security.Puts[put].LastTradeDays >= 2 {
		lastTrade = fmt.Sprintf("%dd", security.Puts[put].LastTradeDays)
	}

	earnings := ""
	if security.EarningsDate != "" && security.EarningsDate <= expiration {
		earnings = "E"
	}
	pricePlus := ""
	if security.Puts[put].Strike >= security.Price {
		pricePlus = "O"
	}
	oddLot := ""
	if security.Puts[put].Size != 100 {
		oddLot = fmt.Sprintf("%d", security.Puts[put].Size)
	}

	fmt.Printf("%8s", security.Ticker)
	fmt.Printf(separator)
	fmt.Printf("%10s", security.Puts[put].Expiration)
	fmt.Printf(separator)
	fmt.Printf("%8.2f", security.Price)
	fmt.Printf(separator)
	fmt.Printf("%8.2f", security.Puts[put].Strike)
	fmt.Printf(separator)
	fmt.Printf("%8.2f", security.Puts[put].Last)
	fmt.Printf(separator)
	fmt.Printf("%8.2f", security.Puts[put].Bid)
	fmt.Printf(separator)
	fmt.Printf("%8.2f", security.Puts[put].Ask)
	fmt.Printf(separator)
	fmt.Printf("%8.1f%%", security.Puts[put].BidStrikeRatio)
	fmt.Printf(separator)
	fmt.Printf("%7.1f%%", security.Puts[put].SafetySpread)
	fmt.Printf(separator)
	fmt.Printf("%8s", lastTrade)
	fmt.Printf(separator)
	fmt.Printf("%8s", earnings)
	fmt.Printf(separator)
	fmt.Printf("%8s", pricePlus)
	fmt.Printf(separator)
	fmt.Printf("%8s", oddLot)
	fmt.Printf("\n")
}
