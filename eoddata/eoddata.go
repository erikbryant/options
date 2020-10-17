package eoddata

import (
	"github.com/erikbryant/options/csv"
	"sort"
	"strings"
	"unicode"
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
		skip := false
		for _, char := range cells[0] {
			if !unicode.IsLetter(char) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		securities = append(securities, ticker)
	}

	sort.Strings(securities)

	return securities, nil
}
