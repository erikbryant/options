package tradeking

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/options/cache"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"
)

var (
	cipherOauthToken  = "zCeloKfLfBLbEaHTTLWJLn86/sOUOpKqrphui9mPKcVz32nxOsapurN+SStCBRUm6TZku/kH2VF7OllDn3fb9G7pei6vD9L4"
	oauthToken        = ""
	cipherConsumerKey = "uA/gknxobBKarpEjTs8lEtC4+Q0tSCnp5755Dcgjv3PLhASbPmLlEWxRS6Cpu0Qs/2J5MR2yGxAyDAR0bGVRmR4sF71B7j8f"
	consumerKey       = ""
)

// Init initializes the internal state of the package.
func Init(passPhrase string) {
	var err error

	oauthToken, err = aes.Decrypt(cipherOauthToken, passPhrase)
	if err != nil {
		panic("Incorrect passphrase for Trade King oauth token")
	}

	consumerKey, err = aes.Decrypt(cipherConsumerKey, passPhrase)
	if err != nil {
		panic("Incorrect passphrase for Trade King oauth consumer key")
	}
}

// parseMarketOptions extracts salient information from the raw Trade King format.
func parseMarketOptions(m map[string]interface{}, security sec.Security) (sec.Security, error) {
	quotes, err := extractQuotes(m)
	if err != nil {
		return security, err
	}

	quote, ok := quotes["quote"]
	if !ok {
		return security, fmt.Errorf("Unable to parse quote")
	}

	strikes := make(map[float64]bool)

	for _, val := range quote.([]interface{}) {
		symbol, ok := val.(map[string]interface{})["symbol"]
		if !ok {
			return security, fmt.Errorf("Unable to parse symbol")
		}

		rootsymbol, ok := val.(map[string]interface{})["rootsymbol"]
		if !ok {
			return security, fmt.Errorf("Unable to parse rootsymbol")
		}
		if rootsymbol != security.Ticker {
			if len(rootsymbol.(string)) <= len(security.Ticker) || rootsymbol.(string)[:len(security.Ticker)] != security.Ticker {
				// Some options have a number prepended to their rootsymbol.
				// ABBV has a rootsymbol of ABBV1. That's a close enough match.
				return security, fmt.Errorf("These options do not match ticker %s %s %s", rootsymbol, security.Ticker, symbol)
			}
		}

		var contract sec.Contract
		var err error

		xdate, ok := val.(map[string]interface{})["xdate"]
		if !ok {
			return security, fmt.Errorf("Unable to parse xdate")
		}
		contract.Expiration = xdate.(string)

		strikeprice, ok := val.(map[string]interface{})["strikeprice"]
		if !ok {
			return security, fmt.Errorf("Unable to parse strikeprice")
		}
		contract.Strike, err = strconv.ParseFloat(strikeprice.(string), 64)
		if err != nil {
			return security, err
		}
		strikes[contract.Strike] = true

		last, ok := val.(map[string]interface{})["last"]
		if !ok {
			return security, fmt.Errorf("Unable to parse last")
		}
		contract.Last, err = strconv.ParseFloat(last.(string), 64)
		if err != nil {
			return security, err
		}

		bid, ok := val.(map[string]interface{})["bid"]
		if !ok {
			return security, fmt.Errorf("Unable to parse bid")
		}
		contract.Bid, err = strconv.ParseFloat(bid.(string), 64)
		if err != nil {
			return security, err
		}

		ask, ok := val.(map[string]interface{})["ask"]
		if !ok {
			return security, fmt.Errorf("Unable to parse ask")
		}
		contract.Ask, err = strconv.ParseFloat(ask.(string), 64)
		if err != nil {
			return security, err
		}

		contractSize, ok := val.(map[string]interface{})["contract_size"]
		if !ok {
			return security, fmt.Errorf("Unable to parse contract_size")
		}
		contract.Size, err = strconv.Atoi(contractSize.(string))
		if err != nil {
			return security, err
		}

		date, ok := val.(map[string]interface{})["date"]
		if !ok {
			return security, fmt.Errorf("Unable to parse date")
		}
		contract.LastTradeDate, err = time.Parse("2006-01-02", date.(string))
		if err != nil {
			return security, err
		}

		putCall, ok := val.(map[string]interface{})["put_call"]
		if !ok {
			return security, fmt.Errorf("Unable to parse put_call")
		}

		if putCall == "put" {
			security.Puts = append(security.Puts, contract)
		} else if putCall == "call" {
			security.Calls = append(security.Calls, contract)
		} else {
			return security, fmt.Errorf("Found contract that was neither put nor call %s", putCall)
		}
	}

	for key := range strikes {
		security.Strikes = append(security.Strikes, key)
	}

	sort.Float64s(security.Strikes)

	return security, nil
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
		return nil, fmt.Errorf("Error fetching security data %s", message)
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
func parseMarketExt(m map[string]interface{}, security sec.Security) (sec.Security, error) {
	quotes, err := extractQuotes(m)
	if err != nil {
		return security, err
	}

	quote, ok := quotes["quote"]
	if !ok {
		return security, fmt.Errorf("Unable to parse quote")
	}

	last, ok := quote.(map[string]interface{})["last"]
	if !ok {
		return security, fmt.Errorf("Unable to parse last")
	}

	security.Price, err = strconv.ParseFloat(last.(string), 64)
	if err != nil {
		return security, err
	}

	return security, nil
}

func webRequest(url string) (map[string]interface{}, bool, error) {
	var response *http.Response
	var err error

	// Oauth 1.0 authentication
	ovr := "&oauth_version=1.0"
	osm := "&oauth_signature_method=HMAC-SHA1"
	ots := fmt.Sprintf("&oauth_timestamp=%d", time.Now().Unix())
	ock := "&oauth_consumer_key=" + consumerKey
	otk := "&oauth_token=" + oauthToken

	auth := ock + osm + ots + otk + ovr

	url += auth

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

// GetOptions looks up a single ticker symbol and returns its options.
func GetOptions(security sec.Security) (sec.Security, error) {
	cacheStale := false

	today := time.Now().Format("20060102")

	url := "https://api.tradeking.com/v1/market/options/search.json?symbol=" + security.Ticker + "&query=xdate-gte:" + today

	response, err := cache.Read(today + url)
	if err != nil {
		cacheStale = true
		for {
			var retryable bool
			response, retryable, err = webRequest(url)
			if retryable {
				continue
			}
			if err != nil {
				return security, fmt.Errorf("Error fetching option data %s %s", security.Ticker, err)
			}
			break
		}
	}

	security, err = parseMarketOptions(response, security)
	if err != nil {
		return security, fmt.Errorf("Error parsing market options %s", err)
	}

	if cacheStale {
		cache.Update(today+url, response)
	}

	return security, nil
}

// GetStock looks up a single ticker symbol returns the security.
func GetStock(security sec.Security) (sec.Security, bool, error) {
	today := time.Now().Format("20060102")

	url := "https://api.tradeking.com/v1/market/ext/quotes.json?symbols=" + security.Ticker

	response, err := cache.Read(today + url)
	if err != nil {
		var retryable bool
		response, retryable, err = webRequest(url)
		if err != nil {
			return security, retryable, fmt.Errorf("Error fetching stock data %s %s", security.Ticker, err)
		}
	}

	security, err = parseMarketExt(response, security)
	if err != nil {
		return security, false, fmt.Errorf("Error parsing market ext %s", err)
	}

	// Only update the cache if the price was populated.
	if security.Price > 0 {
		cache.Update(today+url, response)
	}

	return security, false, nil
}
