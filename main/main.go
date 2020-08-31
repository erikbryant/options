package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/options"
	csvFmt "github.com/erikbryant/options/csv"
	"sort"
	"strings"
	"time"
)

var (
	tickers     = flag.String("tickers", "", "Comma separated list of stocks to get option data for")
	csv         = flag.Bool("csv", false, "Output in CSV format?")
	regenerate  = flag.Bool("regenerate", false, "Regenerate option database?")
	useFile     = flag.String("useFile", "", "USE equity database filename")
	optionsFile = flag.String("optionsFile", "", "Options database filename")
	all         = flag.Bool("all", false, "Use all of the options?")
	expiration  = flag.String("expiration", "", "Only options up to this expiration")
	maxStrike   = flag.Float64("maxStrike", 999999999, "Only tickers below this strike price")
	minYield    = flag.Float64("minYield", 0, "Only tickers with at least this bid/strike yield")
	minSafety   = flag.Float64("minSafety", 0, "Only tickers with at least this safety spread")
	skiplist    = flag.String("skiplist", "LABD,LABU,SQQQ,UVXY", "Comma separated list of stocks to skip")
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  Generate the list of all tickers with options (skipping known options)")
	fmt.Println("    options -regenerate -useFile <USE_nnnnnnnn.csv> -optionsFile options.csv")
	fmt.Println("  Find all option plays")
	fmt.Println("    options -all -optionsFile options.csv -expiration 20200821 [-csv] [-header]")
	fmt.Println("  Find interesting option plays")
	fmt.Println("    options -all -optionsFile options.csv  -expiration 20200821[-csv] [-header] [-maxStrike 32.20] [-minYield 4.5]")
	fmt.Println("  Find interesting option plays, limited to these tickers")
	fmt.Println("    options -tickers=<ticker1,ticker2,...>  -expiration 20200821[-csv] [-header] [-maxStrike 32.20] [-minYield 4.5]")
}

// combine merges two lists into one, removes any elements that are in skip, and returns the sorted remainder.
func combine(list1, list2 []string, skip []string) []string {
	m := make(map[string]int)

	for _, val := range list1 {
		m[val] = 1
	}

	for _, val := range list2 {
		m[val] = 1
	}

	for _, val := range skip {
		delete(m, val)
	}

	var result []string

	for key := range m {
		if key == "" {
			continue
		}
		result = append(result, key)
	}

	sort.Strings(result)

	return result
}

func main() {
	flag.Parse()

	err := options.Init(time.Now().Format("20060102"), *expiration)
	if err != nil {
		fmt.Println(err)
		return
	}

	if *regenerate {
		_, err := options.FindSecuritiesWithOptions(*useFile, *optionsFile)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	if *tickers == "" && !*all {
		fmt.Println("You must specify at least one of '-all' or '-tickers'")
		usage()
		return
	}

	var t []string

	if *all {
		t, err = csvFmt.GetFile(*optionsFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	t = combine(t, strings.Split(*tickers, ","), strings.Split(*skiplist, ","))

	securities, err := options.Securities(t)
	if err != nil {
		fmt.Println("Error getting security data", err)
		return
	}

	header := true
	for _, security := range securities {
		for put := range security.Puts {
			if security.Puts[put].Strike > *maxStrike {
				continue
			}

			if *expiration != security.Puts[put].Expiration {
				continue
			}

			if security.Puts[put].Bid <= 0 {
				continue
			}

			if security.Puts[put].BidStrikeRatio < *minYield {
				continue
			}

			if security.Puts[put].SafetySpread < *minSafety {
				continue
			}

			// Does this option cost more than current market share price?
			if security.Puts[put].PriceBasisDelta <= 0 {
				// Sometimes in-the-money strikes have such high bids that
				// they result in a cost basis below the current price.
				continue
			}

			security.PrintPut(put, *csv, header, *expiration)

			header = false
		}
	}
}
