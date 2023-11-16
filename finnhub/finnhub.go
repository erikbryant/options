package finnhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/erikbryant/aes"
	"github.com/erikbryant/options/cache"
	"github.com/erikbryant/options/date"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
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
			return nil, false, fmt.Errorf("error fetching symbol data %s", err)
		}
		if response.StatusCode == 429 {
			reset, err := strconv.ParseInt(response.Header["X-Ratelimit-Reset"][0], 10, 64)
			if err != nil {
				return nil, true, fmt.Errorf("Throttled")
			}
			seconds := reset - time.Now().Unix()
			if seconds <= 0 || seconds > 10 {
				return nil, true, fmt.Errorf("Throttled")
			}
			after, err := time.ParseDuration(fmt.Sprintf("%ds", seconds))
			if err != nil || after < 5 {
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

// parseEarnings parses the earnings json returned from finnhub.
func parseEarnings(m map[string]interface{}) (map[string]string, error) {
	earnings := make(map[string]string)

	earningsCalendar, ok := m["earningsCalendar"]
	if !ok {
		return nil, fmt.Errorf("unable to parse earningsCalendar object")
	}

	for _, val := range earningsCalendar.([]interface{}) {
		symbol, ok := val.(map[string]interface{})["symbol"]
		if !ok {
			return nil, fmt.Errorf("unable to parse symbol object")
		}

		date, ok := val.(map[string]interface{})["date"]
		if !ok {
			return nil, fmt.Errorf("unable to parse date object")
		}

		earnings[symbol.(string)] = date.(string)
	}

	return earnings, nil
}

// parseQuote parses the quote json returned from finnhub.
func parseQuote(m map[string]interface{}, sec security.Security) (security.Security, error) {
	t, ok := m["t"]
	if !ok {
		return sec, fmt.Errorf("unable to parse quote object timestamp")
	}

	c, ok := m["c"]
	if !ok {
		return sec, fmt.Errorf("unable to parse quote object close")
	}

	sec.Price, ok = c.(float64)
	if !ok {
		return sec, fmt.Errorf("unable to convert c to float64 %v", c)
	}

	now := time.Now()
	quoteDate := time.Unix(int64(t.(float64)), 0)
	sinceClose := date.TimeSinceClose(now)
	if now.Sub(quoteDate) > (sinceClose + 6*time.Hour + 30*time.Minute) {
		return sec, fmt.Errorf("security price is stale %s %f %d %v %v %v", sec.Ticker, sec.Price, int64(t.(float64)), quoteDate, sinceClose, now.Sub(quoteDate))
	}

	return sec, nil
}

// parseQuote parses the quote json returned from finnhub.
func parseMetric(m map[string]interface{}, sec security.Security) (security.Security, error) {
	mMetric, ok := m["metric"]
	if !ok {
		return sec, fmt.Errorf("unable to parse metric object")
	}

	metric, ok := mMetric.(map[string]interface{})
	if !ok {
		return sec, fmt.Errorf("unable to decode metric object")
	}

	pe, ok := metric["peBasicExclExtraTTM"]
	if !ok {
		return sec, fmt.Errorf("unable to parse metric object peBasicExclExtraTTM")
	}

	if pe != nil {
		sec.PE, ok = pe.(float64)
		if !ok {
			return sec, fmt.Errorf("unable to convert pe to float64 %v", pe)
		}
	}

	return sec, nil
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
				return nil, fmt.Errorf("error fetching earnings dates %s", err)
			}
			break
		}
	}

	dates, err := parseEarnings(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing earning dates %s", err)
	}

	cache.Update(today+url, response)

	return dates, nil
}

func getQuote(sec security.Security) (security.Security, bool, error) {
	cacheStale := false
	today := time.Now().Format("20060102")

	url := "https://finnhub.io/api/v1/quote?symbol=" + sec.Ticker

	response, err := cache.Read(today + url)
	if err != nil {
		cacheStale = true
		var retryable bool
		response, retryable, err = webRequest(url)
		if err != nil {
			return sec, retryable, fmt.Errorf("error fetching quote data %s %s", sec.Ticker, err)
		}
	}

	sec, err = parseQuote(response, sec)
	if err != nil {
		return sec, false, fmt.Errorf("error parsing quote %s", err)
	}

	// Only update the cache if the price was populated.
	if cacheStale && sec.Price > 0 {
		cache.Update(today+url, response)
	}

	return sec, false, nil
}

func getMetric(sec security.Security) (security.Security, bool, error) {
	cacheStale := false
	today := time.Now().Format("20060102")

	url := "https://finnhub.io/api/v1/stock/metric?metric=all&symbol=" + sec.Ticker

	response, err := cache.Read(today + url)
	if err != nil {
		cacheStale = true
		var retryable bool
		response, retryable, err = webRequest(url)
		if err != nil {
			return sec, retryable, fmt.Errorf("error fetching metric data %s %s", sec.Ticker, err)
		}
	}

	sec, err = parseMetric(response, sec)
	if err != nil {
		return sec, false, fmt.Errorf("error parsing metric %s", err)
	}

	// Only update the cache if the price was populated.
	if cacheStale && sec.Price > 0 {
		cache.Update(today+url, response)
	}

	return sec, false, nil
}

// GetStock looks up a single ticker symbol returns the sec.
func GetStock(sec security.Security) (security.Security, bool, error) {
	sec, retryable, err := getQuote(sec)
	if err != nil {
		return sec, retryable, err
	}

	sec, retryable, err = getMetric(sec)
	return sec, retryable, err
}
