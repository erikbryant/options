package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/erikbryant/options/yahoo"
	"github.com/erikbryant/web"
	"regexp"
	"strings"
)

var (
	tickers = flag.String("tickers", "", "Comma separated list of stocks to get option data for")
	csv     = flag.Bool("csv", false, "Output in CSV format?")
	header  = flag.Bool("header", true, "Write header line?")
)

func main() {
	flag.Parse()

	if *tickers == "" {
		fmt.Println("You must specify `-tickers=<ticker1,ticker2,...>`")
		return
	}

	for _, ticker := range strings.Split(*tickers, ",") {
		optionContractsStore, err := yahoo.Symbol(ticker)
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
		printTicker(ticker, sec, put, *csv, *header)
		*header = false
	}
}
