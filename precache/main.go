package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/options/date"
	"github.com/erikbryant/options/skiplist"
	"github.com/erikbryant/options/utils"
	"time"

	"github.com/erikbryant/options/cboe"
	"github.com/erikbryant/options/marketData"
)

var (
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock API key(s)")
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  Precache files so the Friday night run does not exceed our daily quota")
	fmt.Println("    precache -passPhrase XYZZY")
}

func main() {
	flag.Parse()

	if *passPhrase == "" {
		fmt.Println("You must specify a passPhrase")
		usage()
		return
	}

	marketData.Init(*passPhrase)

	// Construct the list of options to scan
	weekly, err := cboe.WeeklyOptions()
	if err != nil {
		fmt.Printf("error loading CBOE weekly options list %s\n", err)
		return
	}
	weekly = utils.Remove(weekly, skiplist.Skip)

	// Precache the candles from last week
	startDate := date.Previous(time.Monday)
	fmt.Printf("Using startDate: %s for previous week's price %%change\n\n", startDate)
	for _, symbol := range weekly {
		fmt.Printf("\r%s    ", symbol)
		_, _, err = marketData.Candle(symbol, startDate)
		if err != nil {
			fmt.Printf("error getting candle: %s %s\n", startDate, err)
		}
	}
	fmt.Println()
}
