package finnhub

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/options/cache"
	"github.com/erikbryant/web"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func webRequest(url string) (map[string]interface{}, error) {
	var response *http.Response
	var err error

	// API key authentication
	auth := "&token="

	url += auth

	for {
		response, err = web.Request2(url, map[string]string{})
		if err != nil {
			return nil, fmt.Errorf("Error fetching symbol data %s", err)
		}
		if response.StatusCode == 429 {
			after, err := time.ParseDuration(response.Header["X-Ratelimit-Retry-After"][0] + "s")
			if err != nil {
				// Not sure why it would fail, but sleep for at least a little while.
				after, _ = time.ParseDuration("5s")
			}
			fmt.Printf("Throttled. Backing off for %s...", after)
			time.Sleep(after)
			fmt.Printf("done\n")
			continue
		}
		if response.StatusCode == 200 {
			break
		}
		return nil, fmt.Errorf("Got an unexpected StatusCode %v", response)
	}

	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(strings.NewReader(string(resp)))
	var m interface{}
	err = dec.Decode(&m)
	if err != nil {
		return nil, err
	}

	// If the web request was successful we should get back a
	// map in JSON form. If it failed we should get back an error
	// message in string form. Make sure we got a map.
	f, ok := m.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("RequestJSON: Expected a map, got: /%s/", string(resp))
	}

	return f, nil
}

// parseEarnings parses the json returned from finnhub.
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

// EarningDates finds all earning announcement dates in a given date range.
func EarningDates(start, end string) (map[string]string, error) {
	today := time.Now().Format("20060102")

	s := start[0:4] + "-" + start[4:6] + "-" + start[6:8]
	e := end[0:4] + "-" + end[4:6] + "-" + end[6:8]

	url := "https://finnhub.io/api/v1/calendar/earnings?from=" + s + "&to=" + e

	response, err := cache.Read(today + url)
	if err != nil {
		response, err = webRequest(url)
		if err != nil {
			return nil, fmt.Errorf("Error fetching earnings dates %s", err)
		}
	}

	dates, err := parseEarnings(response)
	if err != nil {
		return nil, fmt.Errorf("Error parsing earning dates %s", err)
	}

	cache.Update(today+url, response)

	return dates, nil
}
