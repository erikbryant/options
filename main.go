package main

import (
	"flag"
	"fmt"
	csvFmt "github.com/erikbryant/options/csv"
	"github.com/erikbryant/options/finnhub"
	"github.com/erikbryant/options/options"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/options/tiingo"
	"github.com/erikbryant/options/tradeking"
	"github.com/erikbryant/options/utils"
	"os"
	"runtime/pprof"
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
	skip        = flag.String("skip", "", "Comma separated list of stocks to skip")
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock API key(s)")
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  Generate the list of all tickers with options (skipping known options)")
	fmt.Println("    options -regenerate -useFile <USE_nnnnnnnn.csv> -optionsFile options.csv -passPhrase XYZZY")
	fmt.Println("  Find all option plays")
	fmt.Println("    options -all -optionsFile options.csv -expiration 20210219 -passPhrase XYZZY")
	fmt.Println("  Find option plays limited to these tickers")
	fmt.Println("    options -tickers=<ticker1,ticker2,...>  -expiration 20210219 -passPhrase XYZZY")
}

var skipList = []string{
	"ACB",
	"APHA",
	"BRZU",
	"CGC",
	"CRON",
	"ERX",
	"EWW",
	"FAS",
	"IWF",
	"JNUG",
	"LABD",
	"LABU",
	"NUGT",
	"SDS",
	"SLV",
	"SNDL",
	"SPXU",
	"SQQQ",
	"TECS",
	"TLRY",
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
		_, err := options.FindSecuritiesWithOptions(*useFile)
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
	t = utils.Combine(t, strings.Split(*tickers, ","), skip)

	// Load underlying data for all tickers.
	securities, err := options.Securities(t)
	if err != nil {
		fmt.Println("Error getting security data", err)
		return
	}

	security.Print(securities, *expiration)
}
