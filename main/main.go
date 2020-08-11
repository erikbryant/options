package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/options"
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
		security, err := options.GetSecurity(ticker)
		if err != nil {
			fmt.Println("Error getting security data", err)
			continue
		}

		security.PrintPuts(*csv, *header)
		*header = false
	}
}
