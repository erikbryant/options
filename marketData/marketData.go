package marketData

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
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

// Init initializes the internal state of the package.
func Init(passPhrase string) {
	var err error

	authToken, err = aes.Decrypt(cipherAuthToken, passPhrase)
	if err != nil {
		panic("Incorrect passphrase for marketData auth token")
	}
}

// parseMarketOptions extracts salient information from the raw Trade King format.
func parseMarketOptions(m map[string]interface{}, sec security.Security) (security.Security, error) {
	quotes, err := extractQuotes(m)
	if err != nil {
		return sec, err
	}

	quote, ok := quotes["quote"]
	if !ok {
		return sec, fmt.Errorf("Unable to parse quote")
	}

	strikes := make(map[float64]bool)

	for _, val := range quote.([]interface{}) {
		symbol, ok := val.(map[string]interface{})["symbol"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse symbol")
		}

		rootsymbol, ok := val.(map[string]interface{})["rootsymbol"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse rootsymbol")
		}
		if rootsymbol != sec.Ticker {
			if len(rootsymbol.(string)) <= len(sec.Ticker) || rootsymbol.(string)[:len(sec.Ticker)] != sec.Ticker {
				// Some options have a number prepended to their rootsymbol.
				// ABBV has a rootsymbol of ABBV1. That's a close enough match.
				return sec, fmt.Errorf("These options do not match ticker %s %s %s", rootsymbol, sec.Ticker, symbol)
			}
		}

		var contract security.Contract
		var err error

		xdate, ok := val.(map[string]interface{})["xdate"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse xdate")
		}
		contract.Expiration = xdate.(string)

		strikeprice, ok := val.(map[string]interface{})["strikeprice"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse strikeprice")
		}
		contract.Strike, err = strconv.ParseFloat(strikeprice.(string), 64)
		if err != nil {
			return sec, err
		}
		strikes[contract.Strike] = true

		delta, ok := val.(map[string]interface{})["idelta"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse idelta")
		}
		if delta != "" {
			contract.Delta, err = strconv.ParseFloat(delta.(string), 64)
			if err != nil {
				return sec, err
			}
		}

		iv, ok := val.(map[string]interface{})["imp_Volatility"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse imp_Volatility")
		}
		if iv != "" {
			contract.IV, err = strconv.ParseFloat(iv.(string), 64)
			if err != nil {
				return sec, err
			}
		}

		last, ok := val.(map[string]interface{})["last"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse last")
		}
		contract.Last, err = strconv.ParseFloat(last.(string), 64)
		if err != nil {
			return sec, err
		}

		bid, ok := val.(map[string]interface{})["bid"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse bid")
		}
		contract.Bid, err = strconv.ParseFloat(bid.(string), 64)
		if err != nil {
			return sec, err
		}

		ask, ok := val.(map[string]interface{})["ask"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse ask")
		}
		contract.Ask, err = strconv.ParseFloat(ask.(string), 64)
		if err != nil {
			return sec, err
		}

		contractSize, ok := val.(map[string]interface{})["contract_size"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse contract_size")
		}
		contract.Size, err = strconv.Atoi(contractSize.(string))
		if err != nil {
			// Sometimes the size is 'na'. That's OK. Anything else is an error.
			if contractSize.(string) != "na" {
				return sec, err
			}
			contract.Size = -1
		}

		date, ok := val.(map[string]interface{})["date"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse date")
		}
		contract.LastTradeDate, err = time.Parse("2006-01-02", date.(string))
		if err != nil {
			return sec, err
		}

		putCall, ok := val.(map[string]interface{})["put_call"]
		if !ok {
			return sec, fmt.Errorf("Unable to parse put_call")
		}

		if putCall == "put" {
			sec.Puts = append(sec.Puts, contract)
		} else if putCall == "call" {
			sec.Calls = append(sec.Calls, contract)
		} else {
			return sec, fmt.Errorf("Found contract that was neither put nor call %s", putCall)
		}
	}

	for key := range strikes {
		sec.Strikes = append(sec.Strikes, key)
	}

	sort.Float64s(sec.Strikes)

	return sec, nil
}

func extractQuotes(m map[string]interface{}) (map[string]interface{}, error) {
	response, ok := m["response"]
	if !ok {
		return nil, fmt.Errorf("Unable to parse response object")
	}

	message, ok := response.(map[string]interface{})["error"]
	if !ok {
		return nil, fmt.Errorf("Unable to parse error message")
	}
	if message != "Success" {
		return nil, fmt.Errorf("Error fetching sec data %s", message)
	}

	quotes, ok := response.(map[string]interface{})["quotes"]
	if !ok {
		return nil, fmt.Errorf("Unable to parse quotes")
	}

	q, ok := quotes.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("quotes was not a map[string]interface{} %v", quotes)
	}

	return q, nil
}

// parseMarketExt extracts salient information from the raw Trade King format.
func parseMarketExt(m map[string]interface{}, sec security.Security) (security.Security, error) {
	quotes, err := extractQuotes(m)
	if err != nil {
		return sec, err
	}

	quote, ok := quotes["quote"]
	if !ok {
		return sec, fmt.Errorf("Unable to parse quote")
	}

	last, ok := quote.(map[string]interface{})["last"]
	if !ok {
		return sec, fmt.Errorf("Unable to parse last")
	}

	sec.Price, err = strconv.ParseFloat(last.(string), 64)
	if err != nil {
		return sec, err
	}

	pe, ok := quote.(map[string]interface{})["pe"]
	if !ok {
		return sec, fmt.Errorf("Unable to parse pe")
	}

	sec.PE, err = strconv.ParseFloat(pe.(string), 64)
	if err != nil {
		return sec, err
	}

	return sec, nil
}

func webRequest(url string) (map[string]interface{}, bool, error) {
	var response *http.Response
	var err error

	fmt.Println(authToken)
	url += "&token=" + authToken

	fmt.Println(url)

	for {
		response, err = web.Request2(url, map[string]string{})
		if err != nil {
			return nil, false, fmt.Errorf("Error fetching symbol data %s", err)
		}
		if response.StatusCode == 429 {
			retryAfter, ok := response.Header["X-Ratelimit-Retry-After"]
			if !ok {
				return nil, true, fmt.Errorf("Could not parse throttling header")
			}
			if len(retryAfter) <= 0 {
				return nil, true, fmt.Errorf("Could not parse throttling seconds")
			}
			after, err := time.ParseDuration(retryAfter[0] + "s")
			if err != nil || after > 5 {
				return nil, true, fmt.Errorf("Throttled")
			}
			after = time.Duration(5 * time.Second)
			fmt.Printf("Throttled. Backing off for %s...", after)
			time.Sleep(after)
			fmt.Printf("done\n")
			continue
		}
		if response.StatusCode == 200 || response.StatusCode == 203 {
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

func cachedFetch(url string) (map[string]interface{}, error) {
	today := time.Now().Format("20060102")

	response, err := cache.Read(today + url)
	if err == nil {
		return response, nil
	}

	for {
		var retryable bool
		response, retryable, err = webRequest(url)
		if retryable {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("Error concatenating marketData option data %s", err)
		}
		break
	}

	cache.Update(today+url, response)

	return response, nil
}

func fetch(url string) (map[string]interface{}, error) {
	response, err := cachedFetch(url)
	if err != nil {
		return nil, err
	}

	for key := range response {
		fmt.Println(key)
	}

	t := []interface{}{}
	for _, val := range response["results"].([]interface{}) {
		t = append(t, val)
	}

	r := response
	for r["next_url"] != nil {
		fmt.Println("next_url:", r["next_url"])
		r, err = cachedFetch(r["next_url"].(string))
		fmt.Println(r)
		for _, val := range r["results"].([]interface{}) {
			t = append(t, val)
		}
		fmt.Println(".")
	}

	response["results"] = t
	fmt.Println()
	for _, val := range response["results"].([]interface{}) {
		fmt.Println(val)
	}
	fmt.Println()
	return response, nil
}

// GetOptions looks up a single ticker symbol and returns its options.
func GetOptions(sec security.Security, expiration string) (security.Security, error) {
	url := "https://api.marketdata.app/v1/options/chain/" + sec.Ticker + "/"
	url += "?expiration=" + expiration
	url += "&strikeLimit=50"

	response, err := fetch(url)
	if err != nil {
		return sec, fmt.Errorf("Error concatenating marketData options %s %s", sec.Ticker, err)
	}

	sec, err = parseMarketOptions(response, sec)
	if err != nil {
		return sec, fmt.Errorf("Error parsing marketData options %s", err)
	}
	// }

	return sec, nil
}

// GetStock looks up a single ticker symbol returns the sec.
func GetStock(sec security.Security) (security.Security, bool, error) {
	today := time.Now().Format("20060102")

	url := "https://api.tradeking.com/v1/market/ext/quotes.json?symbols=" + sec.Ticker

	response, err := cache.Read(today + url)
	if err != nil {
		var retryable bool
		response, retryable, err = webRequest(url)
		if err != nil {
			return sec, retryable, fmt.Errorf("Error fetching stock data %s %s", sec.Ticker, err)
		}
	}

	sec, err = parseMarketExt(response, sec)
	if err != nil {
		return sec, false, fmt.Errorf("Error parsing market ext %s", err)
	}

	// Only update the cache if the price was populated.
	if sec.Price > 0 {
		cache.Update(today+url, response)
	}

	return sec, false, nil
}
