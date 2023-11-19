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

	// Load underlying data for all options
	securities, err := options.Securities(weekly, *expiration)
	if err != nil {
		fmt.Println("Error getting security data", err)
		return
	}

	params := []security.Params{
		{
			Initials:        "cc",
			MaxStrike:       100.0,
			MinYield:        0.0,
			MinSafetySpread: 0.0,
			MinCallSpread:   0.0,
			MinIfCalled:     0.0,
			Itm:             true,
			CallCols:        []string{"ticker", "expiration", "price", "strike", "last", "bid", "ask", "bidPriceRatio", "ifCalled", "delta", "IV", "safetySpread", "callSpread", "age", "earnings", "pe", "lotSize", "notes", "otmItm", "KellyCriterion", "lots", "premium", "outlay"},
			PutCols:         []string{"ticker", "expiration", "price", "strike", "last", "bid", "ask", "bidStrikeRatio", "delta", "IV", "safetySpread", "callSpread", "age", "earnings", "pe", "lotSize", "notes", "otmItm", "KellyCriterion", "lots", "premium", "exposure"},
		},
		{
			Initials:        "eb",
			MaxStrike:       50.0,
			MinYield:        1.5,
			MinSafetySpread: 10.0,
			MinCallSpread:   20.0,
			MinIfCalled:     0.0,
			Itm:             true,
			CallCols:        []string{"ticker", "price", "strike", "bid", "bidPriceRatio", "ifCalled", "delta", "IV", "safetySpread", "callSpread", "age", "earnings", "pe", "itm", "lotSize", "KellyCriterion", "lots", "outlay", "premium"},
			PutCols:         []string{"ticker", "expiration", "price", "strike", "bid", "bidStrikeRatio", "delta", "IV", "safetySpread", "callSpread", "age", "earnings", "pe", "itm", "lotSize", "KellyCriterion", "lots", "exposure", "premium"},
		},
	}

	// The Google Drive ID of the folder to upload to
	parentID := "1BpXjfOqRaSnpv0peBNzA8GcudX2-KMH3"

	for _, param := range params {
		putsSheet, callsSheet := security.Print(securities, *expiration, param)
		upload(putsSheet, parentID)
		upload(callsSheet, parentID)
	}
}
