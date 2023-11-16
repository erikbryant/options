package tiingo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/erikbryant/aes"
	"github.com/erikbryant/options/cache"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
)

var (
	cipherAuthToken = "Y5o23xwp5gRm+XsCO2z9eRX0qdgaKPD1IrXOLKd7MmBavPo2RnY9iamVoowvsuh6JLpJ6LFKVLKXNWoGjE1x70Hn03Y="
	authToken       = ""
)

// Init initializes the internal state of the package.
func Init(passPhrase string) {
	var err error

	authToken, err = aes.Decrypt(cipherAuthToken, passPhrase)
	if err != nil {
		panic("Incorrect passphrase for Tiingo")
	}
}

func webRequest(url string) (map[string]interface{}, bool, error) {
	var response *http.Response
	var err error

	auth := "&token=" + authToken

	url += auth

	headers := map[string]string{
		"content-type": "application/json",
	}

	for {
		response, err = web.Request2(url, headers)
		if err != nil {
			return nil, false, fmt.Errorf("error fetching symbol data %s", err)
		}
		if response.StatusCode == 429 {
			// TODO: Parse the response to see how long to wait.
			return nil, true, fmt.Errorf("throttled")
		}
		if response.StatusCode == 200 {
			break
		}
		return nil, false, fmt.Errorf("got an unexpected StatusCode %v", response)
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, false, err
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return nil, false, fmt.Errorf("unable to unmarshal json %s", err)
	}

	return jsonObject, false, nil
}

// parsePrices parses the quote json returned from finnhub.
func parsePrices(m map[string]interface{}, sec security.Security) (security.Security, error) {
	close, ok := m["close"]
	if !ok {
		return sec, fmt.Errorf("unable to parse prices object")
	}

	sec.Price, ok = close.(float64)
	if !ok {
		return sec, fmt.Errorf("unable to convert c to float64 %v", close)
	}

	return sec, nil
}

// GetPrices looks up a single ticker symbol and returns its options.
func GetPrices(sec security.Security) (security.Security, error) {
	cacheStale := false
	today := time.Now().Format("20060102")

	d := time.Now().Format("2006-1-2")

	url := "https://api.tiingo.com/tiingo/daily/" + strings.ToLower(sec.Ticker) + "/prices?startDate=" + d + "&endDate=" + d + "&format=json&resampleFreq=monthly"

	response, err := cache.Read(today + url)
	if err != nil {
		cacheStale = true
		for {
			var retryable bool
			response, retryable, err = webRequest(url)
			if retryable {
				fmt.Println("Retrying...")
				continue
			}
			if err != nil {
				return sec, fmt.Errorf("error fetching Tiingo option data %s %s", sec.Ticker, err)
			}
			break
		}
	}

	sec, err = parsePrices(response, sec)
	if err != nil {
		return sec, fmt.Errorf("error parsing market options %s", err)
	}

	// Only update the cache if the options fields were populated.
	if cacheStale && sec.HasOptions() {
		cache.Update(today+url, response)
	}

	return sec, nil
}
