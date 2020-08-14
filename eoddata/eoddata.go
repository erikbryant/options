package eoddata

import (
	"fmt"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
)

// GetSymbols loads all known US equity symbols.
func GetSymbols() (map[string]sec.Security, error) {
	url := "https://www.eoddata.com/data/filedownload.aspx?g=1&sd=20200813&ed=20200813&d=4&p=0&o=d&k=vubvanxsz4"

	response, err := web.Request(url, map[string]string{})
	if err != nil {
		return nil, err
	}

	fmt.Println(response)

	return nil, nil
}
