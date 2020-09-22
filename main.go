package main

import (
	"flag"
	"fmt"
	csvFmt "github.com/erikbryant/options/csv"
	"github.com/erikbryant/options/finnhub"
	"github.com/erikbryant/options/options"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/options/tiingo"
	"github.com/erikbryant/options/tradeking"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
)

var (
	cpuprofile  = flag.String("cpuprofile", "", "Enable profiling and write cpu profile to file")
	tickers     = flag.String("tickers", "", "Comma separated list of stocks to get option data for")
	regenerate  = flag.Bool("regenerate", false, "Regenerate option database?")
	useFile     = flag.String("useFile", "", "USE equity database filename")
	optionsFile = flag.String("optionsFile", "", "Options database filename")
	all         = flag.Bool("all", false, "Use all of the options?")
	expiration  = flag.String("expiration", "", "Only options up to this expiration")
	maxStrike   = flag.Float64("maxStrike", 999999999, "Only tickers below this strike price")
	minYield    = flag.Float64("minYield", 0, "Only tickers with at least this bid/strike yield")
	minSafety   = flag.Float64("minSafety", 0, "Only tickers with at least this safety spread")
	skip        = flag.String("skip", "", "Comma separated list of stocks to skip")
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock API key(s)")
	itm         = flag.Bool("itm", true, "Include in-the-money options?")
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

var skipList = []string{
	"ACB",
	"BRZU",
	"ERX",
	"FAS",
	"IWF",
	"JNUG",
	"LABD",
	"LABU",
	"NUGT",
	"SPXU",
	"SQQQ",
	"TECS",
	"TNA",
	"TQQQ",
	"TZA",
	"UCO",
	"UPRO",
	"UVXY",
	"VIXY",
	"VXX",
	"YINN",
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

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	tiingo.Init(*passPhrase)
	tradeking.Init(*passPhrase)
	finnhub.Init(*passPhrase)

	if *regenerate {
		_, err := options.FindSecuritiesWithOptions(*useFile, *optionsFile)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	err := options.Init(time.Now().Format("20060102"), *expiration)
	if err != nil {
		fmt.Println(err)
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

	// Tickers to skip
	skip := strings.Split(*skip, ",")
	for _, val := range skipList {
		skip = append(skip, val)
	}

	// Get the list of tickers to scan.
	t = combine(t, strings.Split(*tickers, ","), skip)

	// Load underlying data for all tickers.
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

			if *expiration < security.Puts[put].Expiration {
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

			// If the put is in the money, only consider it if '-itm=true'.
			if security.Puts[put].Strike > security.Price && !*itm {
				continue
			}

			// Does this option cost more than current market share price?
			if security.Puts[put].PriceBasisDelta <= 0 {
				// Sometimes in-the-money strikes have such high bids that
				// they result in a cost basis below the current price.
				continue
			}

			security.PrintPut(put, header, *expiration)

			header = false
		}
	}

	sec.PrintFooter(*expiration)
}
