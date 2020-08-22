package security

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Contract holds option data for a single expiration date.
type Contract struct {
	Strike     float64
	Last       float64
	Bid        float64
	Ask        float64
	Expiration string
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
	Ticker  string
	Close   DayRange
	Price   float64
	Strikes []float64
	Puts    []Contract
	Calls   []Contract
}

// GetFile reads in the contents of a file.
func GetFile(file string) ([]string, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s", file)
	}

	securities := strings.Split(string(contents), "\n")

	// Strip trailing blank lines
	for securities[len(securities)-1] == "" {
		securities = securities[:len(securities)-1]
	}

	return securities, nil
}

// otmPutStrike finds the nearest put strike that is {outof|at}-the-money.
func (security *Security) otmPutStrike() (float64, error) {
	if len(security.Strikes) == 0 {
		return -1, fmt.Errorf("There are no strikes %v", security)
	}

	for i, strike := range security.Strikes {
		if strike == security.Price {
			return strike, nil
		}
		if strike > security.Price {
			if i > 0 {
				return security.Strikes[i-1], nil
			}
			return -1, fmt.Errorf("Price is lower than lowest strike %v", security)
		}
	}

	return security.Strikes[len(security.Strikes)-1], nil
}

// itmPutStrike finds the nearest put strike that is in-the-money.
func (security *Security) itmPutStrike() (float64, error) {
	for _, strike := range security.Strikes {
		if strike > security.Price {
			return strike, nil
		}
	}

	return -1, fmt.Errorf("Could not find an in-the-money strike %v", security)
}

func (security *Security) getPutForStrike(strike float64) (int, error) {
	for i, put := range security.Puts {
		if put.Strike == strike {
			return i, nil
		}
	}

	return -1, fmt.Errorf("Could not find put for strike %s %f", security.Ticker, strike)
}

// PrintPuts prints the {out|at,in}-the-money option data for a single ticker in CSV or tabulated and with or without a header.
func (security *Security) PrintPuts(csv, header bool) {
	var separator string

	if csv {
		separator = ","
	} else {
		separator = "  "
	}

	if header {
		if csv {
			fmt.Println("share name,share price,expiration,strike,last,bid,ask,b/s ratio,url")
		} else {
			fmt.Printf("%8s", "Ticker")
			fmt.Printf(separator)
			fmt.Printf("%8s", "Price")
			fmt.Printf(separator)
			fmt.Printf("%10s", "Expiration")
			fmt.Printf(separator)
			fmt.Printf("%8s", "Strike")
			fmt.Printf(separator)
			fmt.Printf("%8s", "Last")
			fmt.Printf(separator)
			fmt.Printf("%8s", "Bid")
			fmt.Printf(separator)
			fmt.Printf("%8s", "Ask")
			fmt.Printf(separator)
			fmt.Printf("%5s", "B/S ratio")
			fmt.Printf(separator)
			fmt.Printf("%s", "URL")
			fmt.Printf("\n")
		}
	}

	// Print two strikes (if they exist)
	//   {at|outof}-the-money
	//   in-the-money

	var strikes = []float64{}

	otm, err := security.otmPutStrike()
	if err == nil {
		strikes = append(strikes, otm)
	}

	itm, err := security.itmPutStrike()
	if err == nil {
		strikes = append(strikes, itm)
	}

	for _, strike := range strikes {
		put, err := security.getPutForStrike(strike)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		bsRatio := security.Puts[put].Bid / security.Puts[put].Strike * 100

		if bsRatio < 4.0 {
			continue
		}

		if security.Puts[put].Strike > security.Price {
			// Sometimes in-the-money strikes have such high bids that
			// they result in a cost basis below the current price.
			if !(security.Puts[put].Bid > (security.Puts[put].Strike - security.Price)) {
				// If not, they are not interesting.
				continue
			}
		}

		url := "https://snapshot.fidelity.com/fidresearch/snapshot/landing.jhtml#/research?symbol=" + security.Ticker

		fmt.Printf("%8s", security.Ticker)
		fmt.Printf(separator)
		fmt.Printf("%8.2f", security.Price)
		fmt.Printf(separator)
		fmt.Printf("%10s", security.Puts[put].Expiration)
		fmt.Printf(separator)
		fmt.Printf("%8.2f", security.Puts[put].Strike)
		fmt.Printf(separator)
		fmt.Printf("%8.2f", security.Puts[put].Last)
		fmt.Printf(separator)
		fmt.Printf("%8.2f", security.Puts[put].Bid)
		fmt.Printf(separator)
		fmt.Printf("%8.2f", security.Puts[put].Ask)
		fmt.Printf(separator)
		fmt.Printf("%8.1f%%", bsRatio)
		fmt.Printf(separator)
		fmt.Printf(url)
		fmt.Printf("\n")
	}
}
