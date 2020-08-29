package eoddata

import (
	sec "github.com/erikbryant/options/security"
	"strings"
)

// USEquities loads all known US equity symbols.
func USEquities(useFile string) (map[string]sec.Security, error) {
	equities, err := sec.GetFile(useFile)
	if err != nil {
		return nil, err
	}

	securities := make(map[string]sec.Security)

	for _, equity := range equities[1:] {
		cells := strings.Split(equity, ",")
		var security sec.Security
		security.Ticker = cells[0]
		securities[cells[0]] = security
	}

	return securities, nil
}
