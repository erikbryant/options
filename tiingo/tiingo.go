package tiingo

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/options/cache"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func webRequest(url string) (map[string]interface{}, bool, error) {
	var response *http.Response
	var err error

	auth := "&token="

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
			fmt.Println(response)
			resp, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, false, err
			}
			dec := json.NewDecoder(strings.NewReader(string(resp)))
			var m interface{}
			err = dec.Decode(&m)
			if err != nil {
				return nil, false, err
			}
			fmt.Println(m)
			// after, err := time.ParseDuration(response.Header["X-Ratelimit-Retry-After"][0] + "s")
			// if err != nil || after > 5 {
			// 	return nil, true, fmt.Errorf("Throttled")
			// }
			// fmt.Printf("Throttled. Backing off for %s...", after)
			// time.Sleep(after)
			// fmt.Printf("done\n")
			continue
		}
		if response.StatusCode == 200 {
			break
		}
		return nil, false, fmt.Errorf("Got an unexpected StatusCode %v", response)
	}

	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, false, err
	}

	dec := json.NewDecoder(strings.NewReader(string(resp)))
	var m interface{}
	err = dec.Decode(&m)
	if err != nil {
		return nil, false, err
	}

	// If the web request was successful we should get back a
	// map in JSON form. If it failed we should get back an error
	// message in string form. Make sure we got a map.
	f, ok := m.([]interface{})[0].(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("RequestJSON: Expected []map, got: /%s/", string(resp))
	}

	return f, false, nil
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
	today := time.Now().Format("20060102")

	d := time.Now().Format("2006-1-2")

	url := "https://api.tiingo.com/tiingo/daily/" + strings.ToLower(security.Ticker) + "/prices?startDate=" + d + "&endDate=" + d + "&format=json&resampleFreq=monthly"

	response, err := cache.Read(today + url)
	if err != nil {
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
	// if security.HasOptions() {
	// 	cache.Update(today+url, response)
	// }

	return security, nil
}
