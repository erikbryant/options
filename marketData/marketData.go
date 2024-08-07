package marketData

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/erikbryant/aes"
	"github.com/erikbryant/options/cache"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
)

var (
	cipherAuthToken = "y6b7RDLc/VKHuIiW+IU+ZgIIAmV9IAAlhTDPaiYOifqxVq9O2vKsg+tbjHQZhypeD1drljvv92rP0NH13/VKs8tirx8WzbTAvF1juUpNMg0hnXSq0qWJ"
	authToken       = ""
)

// Init initializes the internal state of the package
func Init(passPhrase string) {
	var err error

	authToken, err = aes.Decrypt(cipherAuthToken, passPhrase)
	if err != nil {
		panic("Incorrect passphrase for marketData auth token")
	}
}

func floats(m map[string]interface{}, key string) ([]float64, error) {
	vals := []float64{}

	data, ok := m[key].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to convert %s", key)
	}

	for _, d := range data {
		v := 0.0
		if d != nil {
			v = d.(float64)
		}
		vals = append(vals, v)
	}

	return vals, nil
}

func int64s(m map[string]interface{}, key string) ([]int64, error) {
	vals := []int64{}

	data, ok := m[key].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to convert %s", key)
	}

	for _, d := range data {
		val, ok := d.(float64)
		if !ok {
			fmt.Printf("Expected a float64, got '%v'. Assuming 0.\n", d)
			val = 0
		}
		vals = append(vals, int64(val))
	}

	return vals, nil
}

// values returns values for the given key
func values(m map[string]interface{}, key string) ([]string, error) {
	vals := []string{}

	data, ok := m[key].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to convert %s", key)
	}

	for _, d := range data {
		vals = append(vals, d.(string))
	}

	return vals, nil
}

// parseMarketOptions extracts salient information from the raw format
func parseMarketOptions(m map[string]interface{}, sec *security.Security) error {
	underlyingPrice, err := floats(m, "underlyingPrice")
	if err != nil {
		return err
	}

	strike, err := floats(m, "strike")
	if err != nil {
		return err
	}

	bid, err := floats(m, "bid")
	if err != nil {
		return err
	}

	ask, err := floats(m, "ask")
	if err != nil {
		return err
	}

	expiration, err := int64s(m, "expiration")
	if err != nil {
		return err
	}

	delta, err := floats(m, "delta")
	if err != nil {
		return err
	}

	iv, err := floats(m, "iv")
	if err != nil {
		return err
	}

	last, err := floats(m, "last")
	if err != nil {
		return err
	}

	openInterest, err := int64s(m, "openInterest")
	if err != nil {
		return err
	}

	updated, err := int64s(m, "updated")
	if err != nil {
		return err
	}

	sec.Price = underlyingPrice[0]

	side, ok := m["side"].([]interface{})
	if !ok {
		return fmt.Errorf("unable to parse side")
	}
	for i, s := range side {
		contract := security.Contract{}
		contract.Strike = strike[i]
		contract.Bid = bid[i]
		contract.Ask = ask[i]
		contract.Delta = delta[i]
		contract.IV = iv[i]
		contract.OpenInterest = openInterest[i]
		contract.Last = last[i]

		t := time.Unix(expiration[i], 0)
		contract.Expiration = t.Format("2006-01-02")

		contract.LastTradeDate = time.Unix(updated[i], 0)

		switch s.(string) {
		case "call":
			sec.Calls = append(sec.Calls, contract)
		case "put":
			sec.Puts = append(sec.Puts, contract)
		default:
			return fmt.Errorf("unable to parse side type: %s", s.(string))
		}
	}

	return nil
}

func webRequest(url string) (map[string]interface{}, bool, error) {
	var response *http.Response
	var err error

	// If this is the first URL parameter we need a '?' prefix
	if url[len(url)-1] == '/' {
		url += "?"
	} else {
		url += "&"
	}
	url += "token=" + authToken

	for {
		response, err = web.Request2(url, map[string]string{})
		if err != nil {
			return nil, false, fmt.Errorf("error requesting symbol data %s", err)
		}
		if response.StatusCode == 200 || response.StatusCode == 203 {
			break
		}
		if response.StatusCode == 429 {
			limitReset, ok := response.Header["X-Api-Ratelimit-Reset"]
			if !ok {
				return nil, true, fmt.Errorf("could not parse throttling header")
			}
			utime, err := strconv.ParseInt(limitReset[0], 10, 64)
			if err != nil {
				return nil, true, err
			}

			t := time.Unix(utime, 0)
			fmt.Printf("Daily quota reached. Sleeping until it resets at %v\n", t)
			for time.Now().Before(t) {
				// We could sleep until t, but Sleep gets confused if the
				// computer hibernates. Instead, sleep/check regularly.
				fmt.Println("Now: ", time.Now(), " Sleeping for 1 minute intervals until: ", t)
				time.Sleep(1 * time.Minute)
			}
			continue
		}
		if response.StatusCode == 509 {
			// Bandwidth limit exceeded (this is retryable)
			fmt.Println("HTTP 509 - Bandwidth limit exceeded; retrying...")
			fmt.Println(response)
			time.Sleep(6 * time.Second)
			continue
		}
		return nil, false, fmt.Errorf("URL %s got an unexpected StatusCode %v", url, response)
	}

	contents, err := io.ReadAll(response.Body)
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

func fetch(url string) (map[string]interface{}, error) {
	response, err := cache.Read(url)
	if err == nil {
		return response, nil
	}

	retry := true
	for retry {
		response, retry, err = webRequest(url)
	}
	if err != nil {
		return nil, err
	}

	cache.Update(url, response)

	return response, nil
}

// leq returns true if date1 is <= date2 AND date1 is not in the past
func leq(date1, date2 string) bool {
	d1, err := time.Parse("2006-01-02", date1)
	if err != nil {
		panic(err)
	}
	d2, err := time.Parse("2006-01-02", date2)
	if err != nil {
		panic(err)
	}

	return (d1.Before(d2) || d1.Equal(d2)) && d1.After(time.Now())
}

func expirationsUpTo(ticker, latestExpiration string) ([]string, error) {
	url := "https://api.marketdata.app/v1/options/expirations/" + ticker + "/"

	response, err := fetch(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s expirations %s", ticker, err)
	}

	dates, err := values(response, "expirations")
	if err != nil {
		return nil, fmt.Errorf("error parsing %s expirations %s", ticker, err)
	}

	expirations := []string{}

	for _, date := range dates {
		if leq(date, latestExpiration) {
			expirations = append(expirations, date)
		}
	}

	return expirations, nil
}

// GetOptions looks up a single ticker symbol and returns its options
func GetOptions(sec *security.Security, latestExpiration string) error {
	expirations, err := expirationsUpTo(sec.Ticker, latestExpiration)
	if err != nil {
		return fmt.Errorf("error getting %s expirations %s", sec.Ticker, err)
	}

	for _, expiration := range expirations {
		url := "https://api.marketdata.app/v1/options/chain/" + sec.Ticker + "/"
		url += "?expiration=" + expiration
		url += "&strikeLimit=11"

		response, err := fetch(url)
		if err != nil {
			return fmt.Errorf("error fetching marketData options %s %s", sec.Ticker, err)
		}

		err = parseMarketOptions(response, sec)
		if err != nil {
			return fmt.Errorf("error parsing marketData options %s", err)
		}
	}

	return nil
}

// Candle returns open/close candlestick data for a symbol
func Candle(symbol, date string) (float64, float64, error) {
	url := "https://api.marketdata.app/v1/stocks/bulkcandles/D/?symbols=" + symbol + "&date=" + date

	response, err := fetch(url)
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching marketData %s candlesticks %s", symbol, err)
	}

	// response = map[
	//   s:ok
	//   symbol:[IBM]
	//   o:[176.15]
	//   h:[176.9]
	//   l:[173.79]
	//   c:[175.1]
	//   t:[1.7095284e+09]
	//   v:[8.1505451e+07]
	// ]

	s := response["s"].(string)
	if s != "ok" {
		return 0, 0, fmt.Errorf("marketData returned non-ok status '%s' for %s", s, symbol)
	}

	o := response["o"].([]interface{})
	c := response["c"].([]interface{})

	return o[0].(float64), c[0].(float64), nil
}

// PctChange returns the %change in share price between open on startDate and close on endDate
func PctChange(symbol, startDate, endDate string) (float64, error) {
	o, _, err := Candle(symbol, startDate)
	if err != nil {
		return 0, err
	}
	_, c, err := Candle(symbol, endDate)
	if err != nil {
		return 0, err
	}
	return 100.0 * (c - o) / o, nil
}
