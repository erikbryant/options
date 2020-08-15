package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/options"
	"strings"
)

var (
	tickers     = flag.String("tickers", "", "Comma separated list of stocks to get option data for")
	csv         = flag.Bool("csv", false, "Output in CSV format?")
	header      = flag.Bool("header", true, "Write header line?")
	regenerate  = flag.Bool("regenerate", false, "Regenerate option database?")
	useFile     = flag.String("useFile", "", "USE equity database filename")
	optionsFile = flag.String("optionsFile", "", "Options database filename")
	all         = flag.Bool("all", false, "Use all of the options?")
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  options -tickers=<ticker1,ticker2,...> [-csv] [-header]")
	fmt.Println("  options -all [-csv] [-header]")
	fmt.Println("  options -regenerate -useFile <USE_nnnnnnnn.csv>")
}

func main() {
	flag.Parse()

	if *regenerate {
		_, err := options.FindSecuritiesWithOptions(*useFile)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	var t []string
	var err error

	if *all {
		t, err = options.SecuritiesWithOptions(*optionsFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if *tickers != "" {
		t = strings.Split(*tickers, ",")
	} else {
		usage()
		return
	}

	for _, ticker := range t {
		security, err := options.Security(ticker)
		if err != nil {
			fmt.Println("Error getting security data", err)
			continue
		}

		security.PrintPuts(*csv, *header)
		*header = false
	}
}
