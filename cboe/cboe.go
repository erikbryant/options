package cboe

import (
	"fmt"
	"github.com/erikbryant/web"
	"github.com/extrame/xls"
	"io"
	"os"
	"sort"
	"unicode"
)

// getData downloads the CBOE weekly options data and saves it to a file.
func getData(file string) error {
	url := "https://www.cboe.com/publish/weelkysmf/weeklysmf.xls"
	headers := map[string]string{}

	resp, err := web.Request2(url, headers)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("Unable to create CBOE data file %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to write to CBOE data file %v", err)
	}

	return nil
}

// WeeklyOptions pulls the list of options with weekly (or more frequent) expirations from the CBOE.
func WeeklyOptions() ([]string, error) {
	file := "web-request-cache/weeklysmf.xls"

	err := getData(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to download CBOE data %v", err)
	}

	w, err := xls.Open(file, "utf-8")
	if err != nil {
		return nil, fmt.Errorf("Unable to open CBOE data %v", err)
	}

	var tickers []string

	sheet := w.GetSheet(0)
	if sheet == nil {
		return nil, fmt.Errorf("Sheet zero not found in CBOE data")
	}

	for i := 0; i <= (int(sheet.MaxRow)); i++ {
		col1 := sheet.Row(i).Col(0)
		var ticker string
		for _, char := range col1 {
			if unicode.IsLetter(char) {
				ticker += string(char)
			}
		}
		if ticker == "Ticker" {
			continue
		}
		if len(ticker) == 0 {
			continue
		}
		tickers = append(tickers, ticker)
	}

	sort.Strings(tickers)

	return tickers, nil
}
