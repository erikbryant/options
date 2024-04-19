package main

import (
	"flag"
	"fmt"

	"github.com/erikbryant/options/cboe"
	"github.com/erikbryant/options/finnhub"
	"github.com/erikbryant/options/gdrive"
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
	fmt.Println("    options -passPhrase XYZZY -expiration 2021-11-19")
}

// Symbols that we do not want to trade in
var skipList = []string{
	// Cannabis
	"ACB",
	"CGC",
	"MSOS",
	"SNDL",
	"TLRY",

	// Leveraged ETFs
	"ERX",
	"FAS",
	"JNUG",
	"LABD",
	"LABU",
	"NUGT",
	"SDS",
	"SLV",
	"SPXU",
	"SQQQ",
	"TNA",
	"TQQQ",
	"UCO",
	"UPRO",
	"UVXY",
	"VIXY",
	"VXX",
	"YINN",

	// MarketData has no options data
	"NANOS",
}

func upload(sheet, parentID string) {
	_, err := gdrive.CreateSheet(sheet, parentID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Uploaded", sheet)
}

func main() {
	flag.Parse()

	if *passPhrase == "" {
		fmt.Println("You must specify a passPhrase")
		usage()
		return
	}

	if *expiration == "" {
		fmt.Println("You must specify an expiration")
		usage()
		return
	}

	finnhub.Init(*passPhrase, *expiration)
	marketData.Init(*passPhrase)

	// Construct the list of options to scan
	weekly, err := cboe.WeeklyOptions()
	if err != nil {
		fmt.Printf("error loading CBOE weekly options list %s\n", err)
		return
	}
	weekly = utils.Remove(weekly, skipList)

	params := []security.Params{
		{
			Initials:        "cc",
			MinPrice:        0.0,
			MaxPrice:        100.0,
			MinYield:        0.0,
			MinSafetySpread: 0.0,
			MinCallSpread:   0.0,
			MinIfCalled:     0.0,
			Itm:             true,
			CallCols:        []string{"ticker", "expiration", "price", "priceChange", "strike", "last", "bid", "ask", "bidPriceRatio", "ifCalled", "delta", "IV", "safetySpread", "callSpread", "age", "earnings", "pe", "lotSize", "notes", "otmItm", "KellyCriterion", "lots", "premium", "outlay"},
			PutCols:         []string{"ticker", "expiration", "price", "priceChange", "strike", "last", "bid", "ask", "bidStrikeRatio", "delta", "IV", "safetySpread", "callSpread", "age", "earnings", "pe", "lotSize", "notes", "otmItm", "KellyCriterion", "lots", "premium", "exposure"},
		},
		{
			Initials:        "eb",
			MinPrice:        2.0,
			MaxPrice:        50.0,
			MinYield:        1.5,
			MinSafetySpread: 10.0,
			MinCallSpread:   20.0,
			MinIfCalled:     0.0,
			Itm:             true,
			CallCols:        []string{"ticker", "price", "priceChange", "strike", "bid", "bidPriceRatio", "ifCalled", "safetySpread", "callSpread", "earnings", "lotSize", "lots", "outlay", "premium"},
			PutCols:         []string{"ticker", "price", "priceChange", "strike", "bid", "bidStrikeRatio", "safetySpread", "callSpread", "earnings", "lotSize", "lots", "exposure", "premium"},
		},
	}

	// Find the max share price we care about; we'll ignore any security above this price
	maxPrice := 0.0
	for _, param := range params {
		if param.MaxPrice > maxPrice {
			maxPrice = param.MaxPrice
		}
	}

	// Load underlying data for all options
	securities, err := options.Securities(weekly, *expiration, maxPrice)
	if err != nil {
		fmt.Println("Error getting security:", err)
		return
	}

	// The Google Drive ID of the folder to upload to
	parentID := "1BpXjfOqRaSnpv0peBNzA8GcudX2-KMH3"

	for _, param := range params {
		putsSheet, callsSheet := security.Print(securities, *expiration, param)
		upload(putsSheet, parentID)
		upload(callsSheet, parentID)
	}
}
