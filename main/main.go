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
	expiration  = flag.String("expiration", "", "Only options with this expiration")
	maxStrike   = flag.Float64("maxStrike", 9999999, "Only tickers below this strike price")
	minYield    = flag.Float64("minYield", 0, "Only tickers with at least this bid/strike yield")
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  Generate the list of all tickers with options")
	fmt.Println("    options -regenerate -useFile <USE_nnnnnnnn.csv>")
	fmt.Println("  Find all option plays")
	fmt.Println("    options -all [-csv] [-header] [-expiration 20200821]")
	fmt.Println("  Find interesting option plays")
	fmt.Println("    options -all [-csv] [-header] [-expiration 20200821] [-maxStrike 32.20] [-minYield 4.5]")
	fmt.Println("  Find interesting option plays, limited to these tickers")
	fmt.Println("    options -tickers=<ticker1,ticker2,...> [-csv] [-header] [-expiration 20200821] [-maxStrike 32.20] [-minYield 4.5]")
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
		security, err := options.Security(ticker, *expiration)
		if err != nil {
			fmt.Println("Error getting security data", err)
			continue
		}

		if security.Price > *maxStrike {
			continue
		}

		security.PrintPuts(*csv, *header)
		*header = false
	}
}
