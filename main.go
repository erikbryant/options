package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/erikbryant/options/cboe"
	"github.com/erikbryant/options/finnhub"
	"github.com/erikbryant/options/marketData"
	"github.com/erikbryant/options/options"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/options/utils"
)

var (
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock API key(s)")
	expiration = flag.String("expiration", "", "Only options up to this expiration")
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  Find all option plays")
	fmt.Println("    options -passPhrase XYZZY -expiration 20211119")
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
	"UCO",
	"UPRO",
	"UVXY",
	"VIXY",
	"VXX",
	"YINN",
}

func main() {
	flag.Parse()

	if *passPhrase == "" {
		fmt.Println("You must specify a passPhrase")
		usage()
		return
	}

	finnhub.Init(*passPhrase)
	marketData.Init(*passPhrase)

	if *expiration == "" {
		fmt.Println("You must specify an expiration")
		usage()
		return
	}

	err := options.Init(time.Now().Format("2006-01-02"), *expiration)
	if err != nil {
		fmt.Println(err)
		return
	}

	t, err := cboe.WeeklyOptions()
	if err != nil {
		fmt.Printf("error loading CBOE weekly options list %s\n", err)
		return
	}

	// Get the list of tickers to scan.
	t = utils.Remove(t, skipList)

	// Load underlying data for all tickers.
	securities, err := options.Securities(t, *expiration)
	if err != nil {
		fmt.Println("Error getting security data", err)
		return
	}

	security.Print(securities, *expiration)
}
