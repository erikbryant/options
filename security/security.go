package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/erikbryant/options/csv"
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
	LotSize       int
	OpenInterest  int64
	Delta         float64
	IV            float64
	// Derived values
	PriceBasisDelta float64 // Share price minus cost basis
	LastTradeDays   int64   // Age of last trade in days
	BidStrikeRatio  float64 // bid / strike
	BidPriceRatio   float64 // bid / strike
	SafetySpread    float64 // distance between share price and cost basis
	CallSpread      float64 // how many strikes out do calls still have bids
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

// Security holds data about a security and its option contracts
type Security struct {
	Ticker       string
	Close        DayRange
	Price        float64 // latest price
	PriceChange  float64 // percent change in price over trailing period
	Puts         []Contract
	Calls        []Contract
	EarningsDate string
	PE           float64
}

// Params holds the parameters for each user's output preferences
type Params struct {
	Initials        string
	MinPrice        float64
	MaxPrice        float64
	MinYield        float64
	MinSafetySpread float64
	MinCallSpread   float64
	MinIfCalled     float64
	Itm             bool
	CallCols        []string
	PutCols         []string
}

// HasOptions returns whether the security has both puts and calls
func (security *Security) HasOptions() bool {
	return len(security.Puts) != 0 && len(security.Calls) != 0
}

// colName returns the column name that a spreadsheet would give it
func colName(cols []string, col string) string {
	for i := range cols {
		if cols[i] == col {
			return fmt.Sprintf("%c", i+'A')
		}
	}

	return "!" + col
}

// CallSpread returns the relative distance to the highest OTM bid that is non-zero
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

// cellPut returns a header and cell string formatted for printing
func (security *Security) cellPut(cols []string, col string, contract Contract, expiration string) (string, string) {
	var h, c string

	switch col {
	case "ticker":
		h = fmt.Sprintf("%8s", "Ticker")
		c = fmt.Sprintf("%8s", security.Ticker)
	case "expiration":
		h = fmt.Sprintf("%10s", "Expiration")
		c = fmt.Sprintf("%10s", contract.Expiration)
	case "price":
		h = fmt.Sprintf("%8s", "Price")
		tickerCol := colName(cols, "ticker")
		c = fmt.Sprintf("\"=googlefinance(%s%d, \"\"price\"\")\"", tickerCol, row)
	case "priceChange":
		h = fmt.Sprintf("%8s", "1wk Price %")
		c = fmt.Sprintf("$%7.02f", security.PriceChange)
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
	case "delta":
		h = fmt.Sprintf("%8s", "Delta")
		c = fmt.Sprintf("%7.1f", contract.Delta)
	case "IV":
		h = fmt.Sprintf("%8s", "IV")
		c = fmt.Sprintf("%7.1f", contract.IV)
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
		// calendar days)
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
	case "pe":
		h = fmt.Sprintf("%8s", "P/E Ratio")
		if security.PE == 0 {
			c = ""
		} else {
			c = fmt.Sprintf("%7.02f", security.PE)
		}
	case "itm":
		h = fmt.Sprintf("%8s", "In the $")
		inTheMoney := ""
		if contract.Strike >= security.Price {
			inTheMoney = "ITM"
		}
		c = fmt.Sprintf("%8s", inTheMoney)
	case "lotSize":
		h = fmt.Sprintf("%8s", "Lot Size")
		c = fmt.Sprintf("%8d", contract.LotSize)
	case "KellyCriterion":
		// Percent of portfolio to risk on a given investment.
		// We use delta as the win factor. Win factors below 0.5
		// result in negative percents. Filter those away.
		// https://www.fidelity.com/viewpoints/active-investor/options-trade-size
		// Puts have negative deltas. We don't want to get assigned
		// on a put, so take the inverse of delta.
		h = "% to Risk"
		deltaCol := colName(cols, "delta")
		c = fmt.Sprintf("\"=if(%s%d = 0, 0, (1+%s%d) - (1-(1+%s%d))/((1+%s%d)/(1-(1+%s%d))))\"", deltaCol, row, deltaCol, row, deltaCol, row, deltaCol, row, deltaCol, row)
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
	default:
		h = fmt.Sprintf("col not found: %s", col)
		c = fmt.Sprintf("col not found: %s", col)
	}

	return h, c
}

// cellCall returns a header and cell string formatted for printing
func (security *Security) cellCall(cols []string, col string, contract Contract, expiration string) (string, string) {
	var h, c string

	switch col {
	case "itm":
		h = fmt.Sprintf("%8s", "In the $")
		inTheMoney := ""
		if contract.Strike <= security.Price {
			inTheMoney = "ITM"
		}
		c = fmt.Sprintf("%8s", inTheMoney)

	case "KellyCriterion":
		// Percent of portfolio to risk on a given investment.
		// We use delta as the win factor. Win factors below 0.5
		// result in negative percents. Filter those away.
		// https://www.fidelity.com/viewpoints/active-investor/options-trade-size
		h = "% to Risk"
		deltaCol := colName(cols, "delta")
		c = fmt.Sprintf("\"=if(%s%d = 0, 0, abs(%s%d) - (1-abs(%s%d))/(abs(%s%d)/(1-abs(%s%d))))\"", deltaCol, row, deltaCol, row, deltaCol, row, deltaCol, row, deltaCol, row)

	default:
		// Everything else is the same for a put as for a call
		h, c = security.cellPut(cols, col, contract, expiration)
	}

	return h, c
}

// formatHeader formats the header for the table
func (security *Security) formatHeader(cols []string) string {
	output := ""
	separator := ","

	// The row with space for the available cash and the week's yield percent
	for _, col := range cols {
		switch col {
		case "premium":
			pname := colName(cols, "premium")
			ename := colName(cols, "exposure")
			if ename[0] == '!' {
				// For calls, the column name is 'outlay'
				ename = colName(cols, "outlay")
			}
			output += fmt.Sprintf("=%s2/%s2", pname, ename)
		}
		output += separator
	}
	output += "\n"

	// The row with the sum of the premium and the exposure
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

	// The column names
	for _, col := range cols {
		h, _ := security.cellPut(cols, col, security.Puts[0], "")
		output += h
		output += separator
	}
	output += "\n"

	return output
}

// formatPut formats the put data for a single ticker
func (security *Security) formatPut(p Params, put int, csv bool, expiration string) string {
	var separator string
	var output string

	if csv {
		separator = ","
	} else {
		separator = "  "
	}

	for _, col := range p.PutCols {
		_, c := security.cellPut(p.PutCols, col, security.Puts[put], expiration)
		output += c
		output += separator
	}
	output += "\n"

	return output
}

// formatCall formats the call data for a single ticker
func (security *Security) formatCall(p Params, call int, csv bool, expiration string) string {
	var separator string
	var output string

	if csv {
		separator = ","
	} else {
		separator = "  "
	}

	for _, col := range p.CallCols {
		_, c := security.cellCall(p.CallCols, col, security.Calls[call], expiration)
		output += c
		output += separator
	}
	output += "\n"

	return output
}

var row = 1

// printPut prints the put data for a single ticker to the personalized CSV files
func (security *Security) printPut(p Params, put int, header bool, expiration string, file string) {
	var output string

	if header {
		row = 1

		output = security.formatHeader(p.PutCols)
		csv.AppendFile(file, output, true)

		row += strings.Count(output, "\n")
	}

	output = security.formatPut(p, put, true, expiration)
	csv.AppendFile(file, output, false)

	row += strings.Count(output, "\n")
}

// printCall prints the call data for a single ticker to the personalized CSV files
func (security *Security) printCall(p Params, call int, header bool, expiration string, file string) {
	var output string

	if header {
		row = 1

		output = security.formatHeader(p.CallCols)
		csv.AppendFile(file, output, true)

		row += strings.Count(output, "\n")
	}

	output = security.formatCall(p, call, true, expiration)
	csv.AppendFile(file, output, false)

	row += strings.Count(output, "\n")
}

func useThisContract(contract Contract, expiration string, p Params) bool {
	if expiration < contract.Expiration {
		return false
	}

	if contract.Bid <= 0 {
		return false
	}

	if contract.Strike < p.MinPrice {
		return false
	}

	if contract.Strike > p.MaxPrice {
		return false
	}

	if contract.SafetySpread < p.MinSafetySpread {
		return false
	}

	if contract.CallSpread < p.MinCallSpread {
		return false
	}

	return true
}

func useThisPut(security Security, contract Contract, expiration string, p Params) bool {
	if !useThisContract(contract, expiration, p) {
		return false
	}

	// Does this option cost more than current market share price?
	if contract.PriceBasisDelta <= 0 {
		return false
	}

	if contract.BidStrikeRatio < p.MinYield {
		return false
	}

	// If it is in the money, only consider it if '-itm=true'.
	if contract.Strike > security.Price && !p.Itm {
		return false
	}

	return true
}

func useThisCall(security Security, contract Contract, expiration string, p Params) bool {
	if !useThisContract(contract, expiration, p) {
		return false
	}

	if contract.BidPriceRatio < p.MinYield {
		return false
	}

	ifCalled := (contract.Bid + contract.Strike - security.Price) / security.Price
	if ifCalled < p.MinIfCalled {
		return false
	}

	// If it is in the money, only consider it if '-itm=true'.
	if contract.Strike < security.Price && !p.Itm {
		return false
	}

	return true
}

// Print writes a filtered set of options to CSV files
func Print(securities []Security, expiration string, p Params) (string, string) {
	putsSheet := p.Initials + "_" + expiration + "_puts.csv"
	header := true
	for _, security := range securities {
		for put, contract := range security.Puts {
			if !useThisPut(security, contract, expiration, p) {
				continue
			}
			security.printPut(p, put, header, expiration, putsSheet)
			header = false
		}
	}

	callsSheet := p.Initials + "_" + expiration + "_calls.csv"
	header = true
	for _, security := range securities {
		for call, contract := range security.Calls {
			if !useThisCall(security, contract, expiration, p) {
				continue
			}
			security.printCall(p, call, header, expiration, callsSheet)
			header = false
		}
	}

	return putsSheet, callsSheet
}
