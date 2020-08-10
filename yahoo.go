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

// prettify formats and prints the input.
func prettify(i interface{}) string {
	s, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Println("Could not Marshal object", i)
	}

	return string(s)
}

// get reads a key from a map[string]interface{} and returns it.
func get(i interface{}, key string) interface{} {
	if i == nil {
		fmt.Println("i is nil trying to get", key)
		return nil
	}
	return i.(map[string]interface{})[key]
}

// price extracts the price of the underlying security.
func price(ocs map[string]interface{}) float64 {
	meta := get(ocs, "meta")
	quote := get(meta, "quote")
	return web.ToFloat64(get(quote, "regularMarketPrice"))
}

// strikes extracts the list of strike prices.
func strikes(ocs map[string]interface{}) []float64 {
	meta := get(ocs, "meta")
	strikes := get(meta, "strikes")

	// Arrays cannot be typecast. Make a copy instead.
	var s []float64
	for _, val := range strikes.([]interface{}) {
		s = append(s, val.(float64))
	}

	return s
}

// otmPutStrike finds the nearest put strike that is out-of-the-money.
func otmPutStrike(ocs map[string]interface{}) float64 {
	price := price(ocs)

	otm := 0.0
	for _, strike := range strikes(ocs) {
		if strike > price {
			return otm
		}
		otm = strike
	}

	return otm
}

//put extracts the information for a single put.
func put(ocs map[string]interface{}, strike float64) (last, bid, ask float64, expiration string, err error) {
	contracts := get(ocs, "contracts")
	puts := get(contracts, "puts")

	for _, val := range puts.([]interface{}) {
		s := get(val, "strike")
		f := get(s, "raw")
		if web.ToFloat64(f) == strike {
			l := get(val, "lastPrice")
			if l == nil {
				return 0, 0, 0, "<not found>", fmt.Errorf("Last is nil")
			}
			last := web.ToFloat64(get(l, "raw"))
			b := get(val, "bid")
			if b == nil {
				return 0, 0, 0, "<not found>", fmt.Errorf("Bid is nil")
			}
			bid := web.ToFloat64(get(b, "raw"))
			a := get(val, "ask")
			if a == nil {
				return 0, 0, 0, "<not found>", fmt.Errorf("Ask is nil")
			}
			ask := web.ToFloat64(get(a, "raw"))
			e := get(val, "expiration")
			expiration := web.ToString(get(e, "fmt"))
			return last, bid, ask, expiration, nil
		}
	}

	return 0, 0, 0, "<not found>", fmt.Errorf("Did not find a put that matched a strike of %f", strike)
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
			return
		}

		strike := otmPutStrike(optionContractsStore)
		last, bid, ask, expiration, err := put(optionContractsStore, strike)
		if err != nil {
			fmt.Println("ERROR: skipping", ticker, err)
			continue
		}

		if *csv {
			fmt.Printf(ticker)
			fmt.Printf(",")
			fmt.Printf("%f", price(optionContractsStore))
			fmt.Printf(",")
			fmt.Printf("%s", expiration)
			fmt.Printf(",")
			fmt.Printf("%f", otmPutStrike(optionContractsStore))
			fmt.Printf(",")
			fmt.Printf("%f", last)
			fmt.Printf(",")
			fmt.Printf("%f", bid)
			fmt.Printf(",")
			fmt.Printf("%f", ask)
			fmt.Printf("\n")
		} else {
			fmt.Printf("%s price: %6.2f\n", ticker, price(optionContractsStore))
			fmt.Printf("   strikes: %v\n", strikes(optionContractsStore))
			fmt.Printf("   OTM put\n")
			fmt.Printf("              expires: %s\n", expiration)
			fmt.Printf("               strike: %6.2f\n", strike)
			fmt.Printf("               last  : %6.2f\n", last)
			fmt.Printf("               bid   : %6.2f\n", bid)
			fmt.Printf("               ask   : %6.2f\n", ask)
			bsRatio := bid / strike * 100
			fmt.Printf("     bid/strike ratio: %6.2f%%\n", bsRatio)
		}
	}
}
