package options

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/options/yahoo"
)

// Contract holds option data for a single expiration date.
type Contract struct {
	Strike     float64
	Last       float64
	Bid        float64
	Ask        float64
	Expiration string
}

// Security holds data about a security and its options.
type Security struct {
	Ticker  string
	Price   float64
	Strikes []float64
	Puts    []Contract
	Calls   []Contract
}

// getRawFloat safely extracts val[key]["raw"] as a float64.
func getRawFloat(i interface{}, key string) (float64, error) {
	val := get(i, key)

	if val == nil {
		return -1, fmt.Errorf("%s was nil", key)
	}

	raw := get(val, "raw")

	if raw == nil {
		return -1, fmt.Errorf("%s[\"raw\"] was nil", key)
	}

	return raw.(float64), nil
}

// getRawFloat safely extracts val[key]["fmt"] as a string.
func getFmtString(i interface{}, key string) (string, error) {
	val := get(i, key)

	if val == nil {
		return "", fmt.Errorf("%s was nil", key)
	}

	f := get(val, "fmt")

	if f == nil {
		return "", fmt.Errorf("%s[\"fmt\"] was nil", key)
	}

	return f.(string), nil
}

// get reads a key from a map[string]interface{} and returns it.
func get(i interface{}, key string) interface{} {
	if i == nil {
		fmt.Println("i is nil trying to get", key)
		return nil
	}
	return i.(map[string]interface{})[key]
}

// parseOCS extracts all of the interesting information from the raw Yahoo! format.
func parseOCS(ocs map[string]interface{}) (Security, error) {
	var sec Security

	// The price of the underlying Security.
	meta := get(ocs, "meta")
	if meta == nil {
		return sec, fmt.Errorf("Meta was nil")
	}

	quote := get(meta, "quote")
	if quote == nil {
		return sec, fmt.Errorf("Quote was nil")
	}

	sec.Price = get(quote, "regularMarketPrice").(float64)

	// The list of strike prices. Arrays cannot be typecast. Make a copy instead.
	strikes := get(meta, "strikes")
	if strikes == nil {
		return sec, fmt.Errorf("Strikes was nil")
	}
	for _, val := range strikes.([]interface{}) {
		if val == nil {
			return sec, fmt.Errorf("Val was nil")
		}
		sec.Strikes = append(sec.Strikes, val.(float64))
	}

	contracts := get(ocs, "contracts")
	if contracts == nil {
		return sec, fmt.Errorf("Contracts was nil")
	}

	// The puts.
	puts := get(contracts, "puts")
	if puts == nil {
		return sec, fmt.Errorf("Puts was nil")
	}

	for _, val := range puts.([]interface{}) {
		var put Contract
		var err error

		put.Strike, err = getRawFloat(val, "strike")
		if err != nil {
			return sec, err
		}

		put.Last, err = getRawFloat(val, "lastPrice")
		if err != nil {
			return sec, err
		}

		put.Bid, err = getRawFloat(val, "bid")
		if err != nil {
			return sec, err
		}

		put.Ask, err = getRawFloat(val, "ask")
		if err != nil {
			return sec, err
		}

		put.Expiration, err = getFmtString(val, "expiration")
		if err != nil {
			return sec, err
		}

		sec.Puts = append(sec.Puts, put)
	}

	return sec, nil
}

// GetSecurity gets all of the relevant data into the Security.
func GetSecurity(ticker string) (Security, error) {
	var sec Security

	optionContractsStore, err := yahoo.Symbol(ticker)
	if err != nil {
		return sec, fmt.Errorf("Error getting security %s %s", ticker, err)
	}

	sec, err = parseOCS(optionContractsStore)
	if err != nil {
		return sec, fmt.Errorf("Error parsing OCS %s", err)
	}

	// TODO: Make Security an object and have parseOCS operate on a Security instead of returning a new one.
	sec.Ticker = ticker

	return sec, nil
}

// prettify formats and prints the input.
func prettify(i interface{}) string {
	s, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Println("Could not Marshal object", i)
	}

	return string(s)
}

// OtmPutStrike finds the nearest put strike that is out-of-the-money.
func OtmPutStrike(sec Security) float64 {
	otm := 0.0
	for _, strike := range sec.Strikes {
		if strike > sec.Price {
			return otm
		}
		otm = strike
	}

	return otm
}

// Put extracts the information for a single put.
func Put(sec Security, strike float64) (int, error) {
	for i, put := range sec.Puts {
		if put.Strike == strike {
			return i, nil
		}
	}

	return -1, fmt.Errorf("Did not find a put that matched a strike of %f", strike)
}

// PrintTicker prints the option data for a single ticker in CSV or tabulated and with or without a header.
func PrintTicker(sec Security, strike float64, put int, csv, header bool) {
	if header {
		if csv {
			fmt.Println("share name,share price,expiration,OTM put strike,last,bid,ask")
		} else {
			fmt.Printf("%8s", "Ticker")
			fmt.Printf("  ")
			fmt.Printf("%8s", "Price")
			fmt.Printf("  ")
			fmt.Printf("%10s", "Expiration")
			fmt.Printf("  ")
			fmt.Printf("%8s", "Strike")
			fmt.Printf("  ")
			fmt.Printf("%8s", "Last")
			fmt.Printf("  ")
			fmt.Printf("%8s", "Bid")
			fmt.Printf("  ")
			fmt.Printf("%8s", "Ask")
			fmt.Printf("  ")
			fmt.Printf("%5s", "B/S ratio")
			fmt.Printf("\n")
		}
	}

	if csv {
		fmt.Printf(sec.Ticker)
		fmt.Printf(",")
		fmt.Printf("%f", sec.Price)
		fmt.Printf(",")
		fmt.Printf("%s", sec.Puts[put].Expiration)
		fmt.Printf(",")
		fmt.Printf("%f", strike)
		fmt.Printf(",")
		fmt.Printf("%f", sec.Puts[put].Last)
		fmt.Printf(",")
		fmt.Printf("%f", sec.Puts[put].Bid)
		fmt.Printf(",")
		fmt.Printf("%f", sec.Puts[put].Ask)
		fmt.Printf("\n")
	} else {
		bsRatio := sec.Puts[put].Bid / sec.Puts[put].Strike * 100
		if bsRatio < 5 {
			return
		}
		fmt.Printf("%8s", sec.Ticker)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.Price)
		fmt.Printf("  ")
		fmt.Printf("%10s", sec.Puts[put].Expiration)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", strike)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.Puts[put].Last)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.Puts[put].Bid)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.Puts[put].Ask)
		fmt.Printf("  ")
		fmt.Printf("%5.1f%%", bsRatio)
		fmt.Printf("\n")
	}
}
