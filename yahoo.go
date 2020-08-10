package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/erikbryant/web"
	"regexp"
	"strings"
)

var (
	tickers = flag.String("tickers", "", "Comma separated list of stocks to get option data for")
	csv     = flag.Bool("csv", false, "Output in CSV format?")
	header  = flag.Bool("header", true, "Write CSV header line?")
)

// symbol looks up a single ticker symbol on Yahoo! Finance and returns the associated JSON data block.
func symbol(s string) (map[string]interface{}, error) {
	url := "https://finance.yahoo.com/quote/" + s + "/options?p=" + s

	response, err := web.Request(url, map[string]string{})
	if err != nil {
		return nil, err
	}

	var re = regexp.MustCompile("root.App.main = ")
	json1 := re.Split(response, 2)
	re = regexp.MustCompile(`;\n}\(this\)\);`)
	json2 := re.Split(json1[1], 2)

	dec := json.NewDecoder(strings.NewReader(string(json2[0])))
	var m interface{}
	err = dec.Decode(&m)
	if err != nil {
		return nil, err
	}

	// If the web request was successful we should get back a
	// map in JSON form. If it failed we should get back an error
	// message in string form. Make sure we got a map.
	f, ok := m.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("RequestJSON: Expected a map, got: /%s/", string(json2[0]))
	}

	context := f["context"]
	dispatcher := context.(map[string]interface{})["dispatcher"]
	stores := dispatcher.(map[string]interface{})["stores"]
	optionContractsStore := stores.(map[string]interface{})["OptionContractsStore"]

	if optionContractsStore == nil {
		return nil, fmt.Errorf("OptionsContractStore is nil")
	}

	return optionContractsStore.(map[string]interface{}), nil
}

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

func main() {
	flag.Parse()

	if *tickers == "" {
		fmt.Println("You must specify `-tickers=<ticker1,ticker2,...>`")
		return
	}

	if *csv && *header {
		fmt.Println("share name,share price,expiration,OTM put strike,last,bid,ask")
	}

	for _, ticker := range strings.Split(*tickers, ",") {
		optionContractsStore, err := symbol(ticker)
		if err != nil {
			fmt.Println("Error requesting symbol data:", ticker, err)
			continue
		}

		sec, err := parseOCS(optionContractsStore)
		if err != nil {
			fmt.Println("Error parsing OCS", ticker, err)
			continue
		}

		strike := otmPutStrike(sec)
		put, err := put(sec, strike)
		if err != nil {
			fmt.Println("Error finding out of the money put", ticker, strike, err)
			continue
		}

		if *csv {
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
			fmt.Printf("%s price: %6.2f\n", ticker, sec.price)
			fmt.Printf("   strikes: %v\n", sec.strikes)
			fmt.Printf("   OTM put\n")
			fmt.Printf("              expires: %s\n", sec.puts[put].expiration)
			fmt.Printf("               strike: %6.2f\n", sec.puts[put].strike)
			fmt.Printf("               last  : %6.2f\n", sec.puts[put].last)
			fmt.Printf("               bid   : %6.2f\n", sec.puts[put].bid)
			fmt.Printf("               ask   : %6.2f\n", sec.puts[put].ask)
			lsRatio := sec.puts[put].last / sec.puts[put].strike * 100
			fmt.Printf("    last/strike ratio: %6.2f%%\n", lsRatio)
		}
	}
}
