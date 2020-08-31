package eoddata

import (
	"github.com/erikbryant/options/csv"
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

		// Skip non-symbols
		if strings.ContainsAny(cells[0], "-.") {
			continue
		}

		// This returns a 500 and does not appear to have options, anyway.
		if cells[0] == "MID" {
			continue
		}

		securities = append(securities, cells[0])
	}

	sort.Strings(securities)

	return securities, nil
}
