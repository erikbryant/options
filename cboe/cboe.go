package cboe

import (
	"fmt"
	"io"
	"strings"

	"github.com/erikbryant/web"
)

// webRequest returns the downloaded payload
func webRequest(url string) ([]byte, error) {
	headers := map[string]string{}

	resp, err := web.Request2(url, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// WeeklyOptions returns options with weekly (or more frequent) expirations from the CBOE
func WeeklyOptions() ([]string, error) {
	url := "https://www.cboe.com/us/options/symboldir/weeklys_options/?download=csv"

	response, err := webRequest(url)
	if err != nil {
		return nil, fmt.Errorf("unable to download CBOE data %v", err)
	}

	var tickers []string

	rows := strings.Split(string(response), "\n")
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if row == "" {
			continue
		}
		cols := strings.Split(row, ",")
		ticker := cols[len(cols)-1]
		ticker = strings.ReplaceAll(ticker, "\"", "")
		tickers = append(tickers, ticker)
	}

	return tickers, nil
}
