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

	header := true
	securities := make(map[string]sec.Security)
	for _, equity := range equities[1:] {
		if header {
			header = false
			continue
		}

		if equity == "" {
			continue
		}

		close := strings.Split(equity, ",")
		var security sec.Security
		security.Ticker = close[0]
		securities[close[0]] = security
	}

	return securities, nil
}
