package cboe

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/erikbryant/options/csv"
	"github.com/erikbryant/web"
)

// getData downloads the CBOE weekly options data and saves it to a file.
func getData(file string) error {
	url := "https://www.cboe.com/us/options/symboldir/weeklys_options/?download=csv"
	headers := map[string]string{}

	resp, err := web.Request2(url, headers)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("unable to create CBOE data file %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("unable to write to CBOE data file %v", err)
	}

	return nil
}

// WeeklyOptions pulls the list of options with weekly (or more frequent) expirations from the CBOE.
func WeeklyOptions() ([]string, error) {
	file := "web-request-cache/cboesymboldirweeklys.csv"

	err := getData(file)
	if err != nil {
		return nil, fmt.Errorf("unable to download CBOE data %v", err)
	}

	records, err := csv.GetFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to open CBOE data %v", err)
	}

	var tickers []string

	for _, record := range records {
		fields := strings.Split(record, "\",\"")
		ticker := strings.Trim(fields[1], "\"")
		tickers = append(tickers, ticker)
	}

	sort.Strings(tickers)

	return tickers, nil
}
