package eoddata

import (
	"github.com/erikbryant/options/csv"
	"github.com/erikbryant/options/utils"
	"sort"
	"strings"
)

// USEquities returns a sorted list of all known US equity symbols.
func USEquities(useFile string) ([]string, error) {
	equities, err := csv.GetFile(useFile)
	if err != nil {
		return nil, err
	}

	var securities []string

	for _, equity := range equities {
		cells := strings.Split(equity, ",")

		ticker := cells[0]

		// Skip non-standard symbols.
		if !utils.AlphaNumeric(ticker) {
			continue
		}

		securities = append(securities, ticker)
	}

	sort.Strings(securities)

	return securities, nil
}
