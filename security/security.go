package security

import (
	"fmt"
	"github.com/erikbryant/options/csv"
	"sort"
	"strings"
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
	BidPriceRatio   float64 // bid / strike
	SafetySpread    float64 // distance between share price and cost basis
	CallSpread      float64 // how many strikes out do calls still have bids
	// Formatted output values
	column map[string]string
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

// params holds the parameters for each user's output preferences.
type params struct {
	maxStrike       float64
	minYield        float64
	minSafetySpread float64
	minCallSpread   float64
	minIfCalled     float64
	itm             bool
	cCols           []string
	pCols           []string
	user            string
}

var (
	paramsEb = params{
		50.0,
		1.5,
		10.0,
		20.0,
		0.0,
		true,
		[]string{"ticker", "price", "strike", "bid", "bidPriceRatio", "ifCalled", "safetySpread", "callSpread", "age", "earnings", "itm", "lotSize", "lots", "outlay", "premium"},
		[]string{"ticker", "expiration", "price", "strike", "bid", "bidStrikeRatio", "safetySpread", "callSpread", "age", "earnings", "itm", "lotSize", "lots", "exposure", "premium"},
		"eb",
	}

	paramsCc = params{
		40.0,
		0.0,
		0.0,
		0.0,
		0.0,
		true,
		[]string{"ticker", "expiration", "price", "strike", "last", "bid", "ask", "bidPriceRatio", "ifCalled", "safetySpread", "callSpread", "age", "earnings", "lotSize", "notes", "otmItm", "lots", "premium", "outlay"},
		[]string{"ticker", "expiration", "price", "strike", "last", "bid", "ask", "bidStrikeRatio", "safetySpread", "callSpread", "age", "earnings", "lotSize", "notes", "otmItm", "lots", "premium", "exposure"},
		"cc",
	}
)

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

// colName returns the column name that a spreadsheet would give it.
func colName(cols []string, col string) string {
	for i := range cols {
		if cols[i] == col {
			return fmt.Sprintf("%c", i+'A')
		}
	}

	return "!" + col
}

// CallSpread returns the relative distance to the highest OTM bid that is non-zero.
func (security *Security) CallSpread(expiration string) float64 {
	maxStrike := 0.0

	for _, call := range security.Calls {
		if call.Expiration != expiration {
			continue
		}
		if call.Strike >= security.Price && call.Bid > 0 && call.Strike > maxStrike {
			maxStrike = call.Strike
		}
	}

	if maxStrike == 0 {
		return 0.0
	}

	return 100.0 * (maxStrike - security.Price) / security.Price
}

// cell returns a header and cell string formatted for printing.
func (security *Security) cell(cols []string, col string, contract Contract, expiration string) (string, string) {
	h := fmt.Sprintf("col not found: %s", col)
	c := fmt.Sprintf("col not found: %s", col)

	switch col {
	case "ticker":
		h = fmt.Sprintf("%8s", "Ticker")
		c = fmt.Sprintf("%8s", security.Ticker)
	case "expiration":
		h = fmt.Sprintf("%10s", "Expiration")
		c = fmt.Sprintf("%10s", contract.Expiration)
	case "price":
		h = fmt.Sprintf("%8s", "Price")
		c = fmt.Sprintf("\"=googlefinance(\"\"%s\"\", \"\"price\"\")\"", security.Ticker)
		// c = fmt.Sprintf("$%7.02f", security.Price)
	case "strike":
		h = fmt.Sprintf("%8s", "Strike")
		c = fmt.Sprintf("$%7.02f", contract.Strike)
	case "last":
		h = fmt.Sprintf("%8s", "Last")
		c = fmt.Sprintf("$%7.02f", contract.Last)
	case "bid":
		h = fmt.Sprintf("%8s", "Bid")
		c = fmt.Sprintf("$%7.02f", contract.Bid)
	case "ask":
		h = fmt.Sprintf("%8s", "Ask")
		c = fmt.Sprintf("$%7.02f", contract.Ask)
	case "bidStrikeRatio":
		h = fmt.Sprintf("%8s", "B/S ratio")
		bidCol := colName(cols, "bid")
		strikeCol := colName(cols, "strike")
		c = fmt.Sprintf("=%s%d/%s%d", bidCol, row, strikeCol, row)
		// c = fmt.Sprintf("%8.1f%%", contract.BidStrikeRatio)
	case "bidPriceRatio":
		h = fmt.Sprintf("%8s", "B/P ratio")
		bidCol := colName(cols, "bid")
		priceCol := colName(cols, "price")
		c = fmt.Sprintf("=%s%d/%s%d", bidCol, row, priceCol, row)
		// c = fmt.Sprintf("%8.1f%%", contract.BidPriceRatio)
	case "ifCalled":
		h = fmt.Sprintf("%8s", "If Called")
		bidCol := colName(cols, "bid")
		priceCol := colName(cols, "price")
		strikeCol := colName(cols, "strike")
		c = fmt.Sprintf("=(%s%d+%s%d-%s%d)/%s%d", bidCol, row, strikeCol, row, priceCol, row, priceCol, row)
	case "safetySpread":
		h = fmt.Sprintf("%8s", "Safety")
		priceCol := colName(cols, "price")
		bidCol := colName(cols, "bid")
		strikeCol := colName(cols, "strike")
		c = fmt.Sprintf("=(%s%d-(%s%d-%s%d))/%s%d", priceCol, row, strikeCol, row, bidCol, row, priceCol, row)
		// c = fmt.Sprintf("%7.1f%%", contract.SafetySpread)
	case "callSpread":
		h = fmt.Sprintf("%8s", "CallSprd")
		c = fmt.Sprintf("%7.1f%%", contract.CallSpread)
	case "age":
		h = fmt.Sprintf("%8s", "Age")
		// TODO: If it is the weekend, then make the threshold for
		// printing higher (i.e., count these as business days, not
		// calendar days).
		var lastTrade string
		if contract.LastTradeDays >= 2 {
			lastTrade = fmt.Sprintf("%dd", contract.LastTradeDays)
		}
		c = fmt.Sprintf("%8s", lastTrade)
	case "earnings":
		h = fmt.Sprintf("%8s", "Earnings")
		earnings := ""
		if security.EarningsDate != "" && security.EarningsDate <= expiration {
			earnings = "E"
		}
		c = fmt.Sprintf("%8s", earnings)
	case "itm":
		h = fmt.Sprintf("%8s", "In the $")
		inTheMoney := ""
		if contract.Strike >= security.Price {
			inTheMoney = "ITM"
		}
		c = fmt.Sprintf("%8s", inTheMoney)
	case "lotSize":
		h = fmt.Sprintf("%8s", "Lot Size")
		c = fmt.Sprintf("%8d", contract.Size)
	case "lots":
		h = fmt.Sprintf("%8s", "Lots")
		c = fmt.Sprintf("%8d", 0)
	case "exposure":
		h = fmt.Sprintf("%8s", "Exposure")
		strikeCol := colName(cols, "strike")
		lotSizeCol := colName(cols, "lotSize")
		lotsCol := colName(cols, "lots")
		c = fmt.Sprintf("=%s%d*%s%d*%s%d", strikeCol, row, lotSizeCol, row, lotsCol, row)
	case "outlay":
		h = fmt.Sprintf("%8s", "Outlay")
		priceCol := colName(cols, "price")
		lotSizeCol := colName(cols, "lotSize")
		lotsCol := colName(cols, "lots")
		c = fmt.Sprintf("=%s%d*%s%d*%s%d", priceCol, row, lotSizeCol, row, lotsCol, row)
	case "premium":
		h = fmt.Sprintf("%8s", "Premium")
		bidCol := colName(cols, "bid")
		lotSizeCol := colName(cols, "lotSize")
		lotsCol := colName(cols, "lots")
		c = fmt.Sprintf("=%s%d*%s%d*%s%d", bidCol, row, lotSizeCol, row, lotsCol, row)
	case "notes":
		h = fmt.Sprintf("%8s", "Notes")
		c = fmt.Sprintf("%8s", "")
	case "otmItm":
		h = fmt.Sprintf("%8s", "OTM/ITM")
		if contract.Strike >= security.Price {
			c = fmt.Sprintf("%8s", "ITM")
		} else {
			c = fmt.Sprintf("%8s", "OTM")
		}
	}

	return h, c
}

// callCell returns a header and cell string formatted for printing.
func (security *Security) callCell(cols []string, col string, contract Contract, expiration string) (string, string) {
	h := fmt.Sprintf("col not found: %s", col)
	c := fmt.Sprintf("col not found: %s", col)

	switch col {
	case "ticker":
		h = fmt.Sprintf("%8s", "Ticker")
		c = fmt.Sprintf("%8s", security.Ticker)
	case "expiration":
		h = fmt.Sprintf("%10s", "Expiration")
		c = fmt.Sprintf("%10s", contract.Expiration)
	case "price":
		h = fmt.Sprintf("%8s", "Price")
		c = fmt.Sprintf("$%7.02f", security.Price)
	case "strike":
		h = fmt.Sprintf("%8s", "Strike")
		c = fmt.Sprintf("$%7.02f", contract.Strike)
	case "last":
		h = fmt.Sprintf("%8s", "Last")
		c = fmt.Sprintf("$%7.02f", contract.Last)
	case "bid":
		h = fmt.Sprintf("%8s", "Bid")
		c = fmt.Sprintf("$%7.02f", contract.Bid)
	case "ask":
		h = fmt.Sprintf("%8s", "Ask")
		c = fmt.Sprintf("$%7.02f", contract.Ask)
	case "bidStrikeRatio":
		h = fmt.Sprintf("%8s", "B/S ratio")
		bidCol := colName(cols, "bid")
		strikeCol := colName(cols, "strike")
		c = fmt.Sprintf("=%s%d/%s%d", bidCol, row, strikeCol, row)
		// c = fmt.Sprintf("%8.1f%%", contract.BidStrikeRatio)
	case "bidPriceRatio":
		h = fmt.Sprintf("%8s", "B/P ratio")
		bidCol := colName(cols, "bid")
		priceCol := colName(cols, "price")
		c = fmt.Sprintf("=%s%d/%s%d", bidCol, row, priceCol, row)
		// c = fmt.Sprintf("%8.1f%%", contract.BidPriceRatio)
	case "ifCalled":
		h = fmt.Sprintf("%8s", "If Called")
		bidCol := colName(cols, "bid")
		priceCol := colName(cols, "price")
		strikeCol := colName(cols, "strike")
		c = fmt.Sprintf("=(%s%d+%s%d-%s%d)/%s%d", bidCol, row, strikeCol, row, priceCol, row, priceCol, row)
	case "safetySpread":
		h = fmt.Sprintf("%8s", "Safety")
		priceCol := colName(cols, "price")
		bidCol := colName(cols, "bid")
		strikeCol := colName(cols, "strike")
		c = fmt.Sprintf("=(%s%d-(%s%d-%s%d))/%s%d", priceCol, row, strikeCol, row, bidCol, row, priceCol, row)
		// c = fmt.Sprintf("%7.1f%%", contract.SafetySpread)
	case "callSpread":
		h = fmt.Sprintf("%8s", "CallSprd")
		c = fmt.Sprintf("%7.1f%%", contract.CallSpread)
	case "age":
		h = fmt.Sprintf("%8s", "Age")
		// TODO: If it is the weekend, then make the threshold for
		// printing higher (i.e., count these as business days, not
		// calendar days).
		var lastTrade string
		if contract.LastTradeDays >= 2 {
			lastTrade = fmt.Sprintf("%dd", contract.LastTradeDays)
		}
		c = fmt.Sprintf("%8s", lastTrade)
	case "earnings":
		h = fmt.Sprintf("%8s", "Earnings")
		earnings := ""
		if security.EarningsDate != "" && security.EarningsDate <= expiration {
			earnings = "E"
		}
		c = fmt.Sprintf("%8s", earnings)
	case "itm":
		h = fmt.Sprintf("%8s", "In the $")
		inTheMoney := ""
		if contract.Strike <= security.Price {
			inTheMoney = "ITM"
		}
		c = fmt.Sprintf("%8s", inTheMoney)
	case "lotSize":
		h = fmt.Sprintf("%8s", "Lot Size")
		c = fmt.Sprintf("%8d", contract.Size)
	case "lots":
		h = fmt.Sprintf("%8s", "Lots")
		c = fmt.Sprintf("%8d", 0)
	case "exposure":
		h = fmt.Sprintf("%8s", "Exposure")
		strikeCol := colName(cols, "strike")
		lotSizeCol := colName(cols, "lotSize")
		lotsCol := colName(cols, "lots")
		c = fmt.Sprintf("=%s%d*%s%d*%s%d", strikeCol, row, lotSizeCol, row, lotsCol, row)
	case "outlay":
		h = fmt.Sprintf("%8s", "Outlay")
		priceCol := colName(cols, "price")
		lotSizeCol := colName(cols, "lotSize")
		lotsCol := colName(cols, "lots")
		c = fmt.Sprintf("=%s%d*%s%d*%s%d", priceCol, row, lotSizeCol, row, lotsCol, row)
	case "premium":
		h = fmt.Sprintf("%8s", "Premium")
		bidCol := colName(cols, "bid")
		lotSizeCol := colName(cols, "lotSize")
		lotsCol := colName(cols, "lots")
		c = fmt.Sprintf("=%s%d*%s%d*%s%d", bidCol, row, lotSizeCol, row, lotsCol, row)
	case "notes":
		h = fmt.Sprintf("%8s", "Notes")
		c = fmt.Sprintf("%8s", "")
	case "otmItm":
		h = fmt.Sprintf("%8s", "OTM/ITM")
		if contract.Strike <= security.Price {
			c = fmt.Sprintf("%8s", "ITM")
		} else {
			c = fmt.Sprintf("%8s", "OTM")
		}
	}

	return h, c
}

// formatHeader formats the header for the table.
func (security *Security) formatHeader(cols []string, csv bool) string {
	var separator string
	var output string

	if csv {
		separator = ","
	} else {
		separator = "  "
	}

	// The row with space for the available cash and the week's yield pct.
	for _, col := range cols {
		switch col {
		case "premium":
			pname := colName(cols, "premium")
			ename := colName(cols, "exposure")
			if ename[0] == '!' {
				// For calls, the column name is 'outlay'.
				ename = colName(cols, "outlay")
			}
			output += fmt.Sprintf("=%s2/%s2", pname, ename)
		}
		output += separator
	}
	output += "\n"

	// The row with the sum of the premium and the exposure.
	for _, col := range cols {
		switch col {
		case "premium":
			name := colName(cols, col)
			output += fmt.Sprintf("=sum(%s4:%s%d)", name, name, 9999)
		case "exposure":
			name := colName(cols, col)
			output += fmt.Sprintf("=sum(%s4:%s%d)", name, name, 9999)
		case "outlay":
			name := colName(cols, col)
			output += fmt.Sprintf("=sum(%s4:%s%d)", name, name, 9999)
		}
		output += separator
	}
	output += "\n"

	// The column names.
	for _, col := range cols {
		h, _ := security.cell(cols, col, security.Puts[0], "")
		output += h
		output += separator
	}
	output += "\n"

	return output
}

// formatPut formats the put data for a single ticker.
func (security *Security) formatPut(p params, put int, csv bool, expiration string) string {
	var separator string
	var output string

	if csv {
		separator = ","
	} else {
		separator = "  "
	}

	for _, col := range p.pCols {
		_, c := security.cell(p.pCols, col, security.Puts[put], expiration)
		output += c
		output += separator
	}
	output += "\n"

	return output
}

// formatCall formats the call data for a single ticker.
func (security *Security) formatCall(p params, call int, csv bool, expiration string) string {
	var separator string
	var output string

	if csv {
		separator = ","
	} else {
		separator = "  "
	}

	for _, col := range p.cCols {
		_, c := security.callCell(p.cCols, col, security.Calls[call], expiration)
		output += c
		output += separator
	}
	output += "\n"

	return output
}

var row = 1

// PrintPut prints the put data for a single ticker to the personalized CSV files.
func (security *Security) PrintPut(p params, put int, header bool, expiration string, file string) {
	var output string

	if header {
		row = 1

		output = security.formatHeader(p.pCols, true)
		csv.AppendFile(file, output, true)

		row += strings.Count(output, "\n")
	}

	output = security.formatPut(p, put, true, expiration)
	csv.AppendFile(file, output, false)

	row += strings.Count(output, "\n")
}

// PrintCall prints the call data for a single ticker to the personalized CSV files.
func (security *Security) PrintCall(p params, call int, header bool, expiration string, file string) {
	var output string

	if header {
		row = 1

		output = security.formatHeader(p.cCols, true)
		csv.AppendFile(file, output, true)

		row += strings.Count(output, "\n")
	}

	output = security.formatCall(p, call, true, expiration)
	csv.AppendFile(file, output, false)

	row += strings.Count(output, "\n")
}

// Print writes a filtered set of options to CSV files.
func Print(securities []Security, expiration string) {
	for _, p := range []params{paramsEb, paramsCc} {
		file := "options_" + p.user + "_puts_" + expiration + ".csv"

		header := true
		for _, security := range securities {
			for put := range security.Puts {
				if expiration < security.Puts[put].Expiration {
					continue
				}

				if security.Puts[put].Bid <= 0 {
					continue
				}

				// Does this option cost more than current market share price?
				if security.Puts[put].PriceBasisDelta <= 0 {
					continue
				}

				if security.Puts[put].Strike > p.maxStrike {
					continue
				}

				if security.Puts[put].BidStrikeRatio < p.minYield {
					continue
				}

				if security.Puts[put].SafetySpread < p.minSafetySpread {
					continue
				}

				if security.Puts[put].CallSpread < p.minCallSpread {
					continue
				}

				// If it is in the money, only consider it if '-itm=true'.
				if security.Puts[put].Strike > security.Price && !p.itm {
					continue
				}

				security.PrintPut(p, put, header, expiration, file)

				header = false
			}
		}

		file = "options_" + p.user + "_calls_" + expiration + ".csv"

		header = true
		for _, security := range securities {
			for call := range security.Calls {
				if expiration < security.Calls[call].Expiration {
					continue
				}

				if security.Calls[call].Bid <= 0 {
					continue
				}

				if security.Calls[call].Strike > p.maxStrike {
					continue
				}

				if security.Calls[call].BidPriceRatio < p.minYield {
					continue
				}

				if security.Calls[call].SafetySpread < p.minSafetySpread {
					continue
				}

				if security.Calls[call].CallSpread < p.minCallSpread {
					continue
				}

				ifCalled := (security.Calls[call].Bid + security.Calls[call].Strike - security.Price) / security.Price
				if ifCalled < p.minIfCalled {
					continue
				}

				// If it is in the money, only consider it if '-itm=true'.
				if security.Calls[call].Strike < security.Price && !p.itm {
					continue
				}

				security.PrintCall(p, call, header, expiration, file)

				header = false
			}
		}
	}
}
