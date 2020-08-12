package security

import (
	"fmt"
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

// getFmtString safely extracts val[key]["fmt"] as a string.
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

// ParseOCS extracts all of the interesting information from the raw Yahoo! format.
func (security *Security) ParseOCS(ocs map[string]interface{}) error {
	// The price of the underlying security.
	meta := get(ocs, "meta")
	if meta == nil {
		return fmt.Errorf("Meta was nil")
	}

	quote := get(meta, "quote")
	if quote == nil {
		return fmt.Errorf("Quote was nil")
	}

	security.Price = get(quote, "regularMarketPrice").(float64)

	// The list of strike prices. Arrays cannot be typecast. Make a copy instead.
	strikes := get(meta, "strikes")
	if strikes == nil {
		return fmt.Errorf("Strikes was nil")
	}
	for _, val := range strikes.([]interface{}) {
		if val == nil {
			return fmt.Errorf("Val was nil")
		}
		security.Strikes = append(security.Strikes, val.(float64))
	}

	contracts := get(ocs, "contracts")
	if contracts == nil {
		return fmt.Errorf("Contracts was nil")
	}

	// The puts.
	puts := get(contracts, "puts")
	if puts == nil {
		return fmt.Errorf("Puts was nil")
	}

	for _, val := range puts.([]interface{}) {
		var put Contract
		var err error

		put.Strike, err = getRawFloat(val, "strike")
		if err != nil {
			return err
		}

		put.Last, err = getRawFloat(val, "lastPrice")
		if err != nil {
			return err
		}

		put.Bid, err = getRawFloat(val, "bid")
		if err != nil {
			return err
		}

		put.Ask, err = getRawFloat(val, "ask")
		if err != nil {
			return err
		}

		put.Expiration, err = getFmtString(val, "expiration")
		if err != nil {
			return err
		}

		security.Puts = append(security.Puts, put)
	}

	return nil
}

// otmPutStrike finds the nearest put strike that is {outof|at}-the-money.
func (security *Security) otmPutStrike() (float64, error) {
	otm := 0.0

	for _, strike := range security.Strikes {
		if strike == security.Price {
			return strike, nil
		}
		if strike > security.Price {
			return otm, nil
		}
		otm = strike
	}

	return -1, fmt.Errorf("Could not find an {outof|at}-the-money strike %v", security)
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
			fmt.Println(err)
			continue
		}

		bsRatio := security.Puts[put].Bid / security.Puts[put].Strike * 100

		if security.Puts[put].Strike > security.Price {
			// Sometimes in-the-money strikes have such high bids that
			// they result in a cost basis below the current price.
			if !(security.Puts[put].Bid > (security.Puts[put].Strike - security.Price)) {
				// If not, they are not interesting.
				continue
			}
		}

		if csv {
			fmt.Printf(security.Ticker)
			fmt.Printf(",")
			fmt.Printf("%f", security.Price)
			fmt.Printf(",")
			fmt.Printf("%s", security.Puts[put].Expiration)
			fmt.Printf(",")
			fmt.Printf("%f", security.Puts[put].Strike)
			fmt.Printf(",")
			fmt.Printf("%f", security.Puts[put].Last)
			fmt.Printf(",")
			fmt.Printf("%f", security.Puts[put].Bid)
			fmt.Printf(",")
			fmt.Printf("%f", security.Puts[put].Ask)
			fmt.Printf("\n")
		} else {
			fmt.Printf("%8s", security.Ticker)
			fmt.Printf("  ")
			fmt.Printf("%8.2f", security.Price)
			fmt.Printf("  ")
			fmt.Printf("%10s", security.Puts[put].Expiration)
			fmt.Printf("  ")
			fmt.Printf("%8.2f", security.Puts[put].Strike)
			fmt.Printf("  ")
			fmt.Printf("%8.2f", security.Puts[put].Last)
			fmt.Printf("  ")
			fmt.Printf("%8.2f", security.Puts[put].Bid)
			fmt.Printf("  ")
			fmt.Printf("%8.2f", security.Puts[put].Ask)
			fmt.Printf("  ")
			fmt.Printf("%5.1f%%", bsRatio)
			fmt.Printf("\n")
		}
	}
}
