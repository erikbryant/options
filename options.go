package options

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/options/yahoo"
	"regexp"
	"strings"
)

type contract struct {
	strike     float64
	last       float64
	bid        float64
	ask        float64
	expiration string
}

type security struct {
	price   float64
	strikes []float64
	puts    []contract
	calls   []contract
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
func parseOCS(ocs map[string]interface{}) (security, error) {
	var sec security

	// The price of the underlying security.
	meta := get(ocs, "meta")
	if meta == nil {
		return sec, fmt.Errorf("Meta was nil")
	}

	quote := get(meta, "quote")
	if quote == nil {
		return sec, fmt.Errorf("Quote was nil")
	}

	sec.price = get(quote, "regularMarketPrice").(float64)

	// The list of strike prices. Arrays cannot be typecast. Make a copy instead.
	strikes := get(meta, "strikes")
	if strikes == nil {
		return sec, fmt.Errorf("Strikes was nil")
	}
	for _, val := range strikes.([]interface{}) {
		if val == nil {
			return sec, fmt.Errorf("Val was nil")
		}
		sec.strikes = append(sec.strikes, val.(float64))
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
		var put contract
		var err error

		put.strike, err = getRawFloat(val, "strike")
		if err != nil {
			return sec, err
		}

		put.last, err = getRawFloat(val, "lastPrice")
		if err != nil {
			return sec, err
		}

		put.bid, err = getRawFloat(val, "bid")
		if err != nil {
			return sec, err
		}

		put.ask, err = getRawFloat(val, "ask")
		if err != nil {
			return sec, err
		}

		put.expiration, err = getFmtString(val, "expiration")
		if err != nil {
			return sec, err
		}

		sec.puts = append(sec.puts, put)
	}

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

// otmPutStrike finds the nearest put strike that is out-of-the-money.
func otmPutStrike(sec security) float64 {
	otm := 0.0
	for _, strike := range sec.strikes {
		if strike > sec.price {
			return otm
		}
		otm = strike
	}

	return otm
}

//put extracts the information for a single put.
func put(sec security, strike float64) (int, error) {
	for i, put := range sec.puts {
		if put.strike == strike {
			return i, nil
		}
	}

	return -1, fmt.Errorf("Did not find a put that matched a strike of %f", strike)
}

func printTicker(ticker string, sec security, put int, csv, header bool) {
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
			fmt.Printf("%5s", "Bid/Strike")
			fmt.Printf("\n")
		}
	}

	if csv {
		fmt.Printf(ticker)
		fmt.Printf(",")
		fmt.Printf("%f", sec.price)
		fmt.Printf(",")
		fmt.Printf("%s", sec.puts[put].expiration)
		fmt.Printf(",")
		fmt.Printf("%f", otmPutStrike(sec))
		fmt.Printf(",")
		fmt.Printf("%f", sec.puts[put].last)
		fmt.Printf(",")
		fmt.Printf("%f", sec.puts[put].bid)
		fmt.Printf(",")
		fmt.Printf("%f", sec.puts[put].ask)
		fmt.Printf("\n")
	} else {
		bsRatio := sec.puts[put].bid / sec.puts[put].strike * 100
		if bsRatio < 5 {
			return
		}
		fmt.Printf("%8s", ticker)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.price)
		fmt.Printf("  ")
		fmt.Printf("%10s", sec.puts[put].expiration)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", otmPutStrike(sec))
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.puts[put].last)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.puts[put].bid)
		fmt.Printf("  ")
		fmt.Printf("%8.2f", sec.puts[put].ask)
		fmt.Printf("  ")
		fmt.Printf("%5.1f%%", bsRatio)
		fmt.Printf("\n")
	}
}
