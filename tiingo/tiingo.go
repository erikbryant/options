package tiingo

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/options/cache"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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
			return nil, false, fmt.Errorf("Error fetching symbol data %s", err)
		}
		if response.StatusCode == 429 {
			// TODO: Parse the response to see how long to wait.
			return nil, true, fmt.Errorf("Throttled")
		}
		if response.StatusCode == 200 {
			break
		}
		return nil, false, fmt.Errorf("Got an unexpected StatusCode %v", response)
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, false, err
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return nil, false, fmt.Errorf("Unable to unmarshal json %s", err)
	}

	return jsonObject, false, nil
}

// parsePrices parses the quote json returned from finnhub.
func parsePrices(m map[string]interface{}, security sec.Security) (sec.Security, error) {
	close, ok := m["close"]
	if !ok {
		return security, fmt.Errorf("Unable to parse prices object")
	}

	security.Price, ok = close.(float64)
	if !ok {
		return security, fmt.Errorf("Unable to convert c to float64 %v", close)
	}

	return security, nil
}

// GetPrices looks up a single ticker symbol and returns its options.
func GetPrices(security sec.Security) (sec.Security, error) {
	cacheStale := false
	today := time.Now().Format("20060102")

	d := time.Now().Format("2006-1-2")

	url := "https://api.tiingo.com/tiingo/daily/" + strings.ToLower(security.Ticker) + "/prices?startDate=" + d + "&endDate=" + d + "&format=json&resampleFreq=monthly"

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
				return security, fmt.Errorf("Error fetching option data %s %s", security.Ticker, err)
			}
			break
		}
	}

	security, err = parsePrices(response, security)
	if err != nil {
		return security, fmt.Errorf("Error parsing market options %s", err)
	}

	// Only update the cache if the options fields were populated.
	if cacheStale && security.HasOptions() {
		cache.Update(today+url, response)
	}

	return security, nil
}
