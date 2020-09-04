package finnhub

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/options/cache"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	cipherAuthToken = "GDrwdFOt/zTS12HUuCYE82Xjdzoa5EYOT9e377XNYc0w2St/CNQ0M/jOZorYzU1G"
	authToken       = ""
)

// Init initializes the internal state of the package.
func Init(passPhrase string) {
	var err error

	authToken, err = aes.Decrypt(cipherAuthToken, passPhrase)
	if err != nil {
		panic("Incorrect passphrase for FinnHub")
	}
}

func webRequest(url string) (map[string]interface{}, bool, error) {
	var response *http.Response
	var err error

	// API key authentication
	auth := "&token=" + authToken

	url += auth

	for {
		response, err = web.Request2(url, map[string]string{})
		if err != nil {
			return nil, false, fmt.Errorf("Error fetching symbol data %s", err)
		}
		if response.StatusCode == 429 {
			reset, err := strconv.ParseInt(response.Header["X-Ratelimit-Reset"][0], 10, 64)
			if err != nil {
				return nil, true, fmt.Errorf("Throttled")
			}
			seconds := reset - time.Now().Unix()
			if seconds <= 0 || seconds > 5 {
				return nil, true, fmt.Errorf("Throttled")
			}
			after, err := time.ParseDuration(fmt.Sprintf("%ds", seconds))
			if err != nil {
				after = time.Duration(5 * time.Second)
			}
			fmt.Printf("Throttled. Backing off for %s...", after)
			time.Sleep(after)
			fmt.Printf("done\n")
			continue
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

// parseEarnings parses the earnings json returned from finnhub.
func parseEarnings(m map[string]interface{}) (map[string]string, error) {
	earnings := make(map[string]string)

	earningsCalendar, ok := m["earningsCalendar"]
	if !ok {
		return nil, fmt.Errorf("Unable to parse earningsCalendar object")
	}

	for _, val := range earningsCalendar.([]interface{}) {
		symbol, ok := val.(map[string]interface{})["symbol"]
		if !ok {
			return nil, fmt.Errorf("Unable to parse symbol object")
		}

		date, ok := val.(map[string]interface{})["date"]
		if !ok {
			return nil, fmt.Errorf("Unable to parse date object")
		}

		earnings[symbol.(string)] = date.(string)
	}

	return earnings, nil
}

// parseQuote parses the quote json returned from finnhub.
func parseQuote(m map[string]interface{}, security sec.Security) (sec.Security, error) {
	c, ok := m["c"]
	if !ok {
		return security, fmt.Errorf("Unable to parse quote object")
	}

	security.Price, ok = c.(float64)
	if !ok {
		return security, fmt.Errorf("Unable to convert c to float64 %v", c)
	}

	return security, nil
}

// EarningDates finds all earning announcement dates in a given date range.
func EarningDates(start, end string) (map[string]string, error) {
	today := time.Now().Format("20060102")

	s := start[0:4] + "-" + start[4:6] + "-" + start[6:8]
	e := end[0:4] + "-" + end[4:6] + "-" + end[6:8]

	url := "https://finnhub.io/api/v1/calendar/earnings?from=" + s + "&to=" + e

	response, err := cache.Read(today + url)
	if err != nil {
		for {
			var retryable bool
			response, retryable, err = webRequest(url)
			if retryable {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("Error fetching earnings dates %s", err)
			}
			break
		}
	}

	dates, err := parseEarnings(response)
	if err != nil {
		return nil, fmt.Errorf("Error parsing earning dates %s", err)
	}

	cache.Update(today+url, response)

	return dates, nil
}

// GetStock looks up a single ticker symbol returns the security.
func GetStock(security sec.Security) (sec.Security, bool, error) {
	cacheStale := false
	today := time.Now().Format("20060102")

	url := "https://finnhub.io/api/v1/quote?symbol=" + security.Ticker

	response, err := cache.Read(today + url)
	if err != nil {
		cacheStale = true
		var retryable bool
		response, retryable, err = webRequest(url)
		if err != nil {
			return security, retryable, fmt.Errorf("Error fetching stock data %s %s", security.Ticker, err)
		}
	}

	security, err = parseQuote(response, security)
	if err != nil {
		return security, false, fmt.Errorf("Error parsing quote %s", err)
	}

	// Only update the cache if the price was populated.
	if cacheStale && security.Price > 0 {
		cache.Update(today+url, response)
	}

	return security, false, nil
}
