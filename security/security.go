package security

import (
	"fmt"
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
		fmt.Printf("%8s", "Note")
		fmt.Printf("\n")
	}

	// TODO: If it is the weekend, then make the threshold for
	// printing higher.
	var lastTrade string
	if security.Puts[put].LastTradeDays >= 2 {
		lastTrade = fmt.Sprintf("%dd", security.Puts[put].LastTradeDays)
	}

	note := ""
	// Earnings
	if security.EarningsDate != "" && security.EarningsDate <= expiration {
		note += "E"
	}
	// Overbidding
	if security.Puts[put].Strike >= security.Price {
		note += " O"
	}
	// Mini options
	if security.Puts[put].Size != 100 {
		note += " M"
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
	fmt.Printf("%8s", note)
	fmt.Printf("\n")
}
