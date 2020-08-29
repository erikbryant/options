package yahoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/erikbryant/options/cache"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
	"io/ioutil"
	"strings"
	"time"
)

// ErrKeyNotFound is an error to use when looking for a key in a map.
var ErrKeyNotFound = errors.New("key was not found")

// getFloat safely extracts a float64 from an interface{}.
func getFloat(i interface{}, key string) (float64, error) {
	val, err := get(i, key)
	if err != nil {
		return -1, err
	}
	if val == nil {
		return 0, fmt.Errorf("key was found, but val is nil: %s", key)
	}

	f, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf("val was found, but is not a float: i[%s] %v", key, val)
	}

	return f, nil
}

// get reads a key from a map[string]interface{} and returns the value.
func get(i interface{}, key string) (interface{}, error) {
	if i == nil {
		return nil, fmt.Errorf("i is nil trying to get: %s", key)
	}

	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("i is not map[string]interface{}: %s", key)
	}

	val, ok := m[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if val == nil {
		return val, nil
	}

	_, ok = val.(interface{})
	if !ok {
		return nil, fmt.Errorf("val was found, but is not interface{}: m[%s] %v", key, val)
	}

	return val, nil
}

// parseContracts extracts the put or call (specified by 'position') options from the OCS.
func parseContracts(ocs map[string]interface{}, position string) ([]sec.Contract, error) {
	// Puts and calls.
	options, err := get(ocs, "options")
	if err != nil {
		return nil, err
	}
	if options == nil {
		return nil, fmt.Errorf("Options was nil")
	}

	var contracts []sec.Contract

	for i, option := range options.([]interface{}) {
		optionsObject, err := get(option, position)
		if err != nil {
			return nil, err
		}
		if optionsObject == nil {
			return nil, fmt.Errorf("Nil value for options[%d][%s]", i, position)
		}

		hasMiniOptions, err := get(ocs, "hasMiniOptions")
		if err != nil {
			return nil, err
		}

		for _, option := range optionsObject.([]interface{}) {
			var contract sec.Contract
			var err error

			hasMO, ok := hasMiniOptions.(bool)
			if ok {
				contract.HasMiniOptions = hasMO
			}

			contract.Strike, err = getFloat(option, "strike")
			if err != nil {
				return nil, err
			}

			contract.Last, err = getFloat(option, "lastPrice")
			if err != nil {
				return nil, err
			}

			contract.Bid, err = getFloat(option, "bid")
			if err != nil {
				// Sometimes this is not present. Without it, the contract is useless.
				if err == ErrKeyNotFound {
					continue
				}
				return nil, err
			}

			contract.Ask, err = getFloat(option, "ask")
			if err != nil {
				// Every now and then ask is not present. That's not fatal.
				if err != ErrKeyNotFound {
					return nil, err
				}
			}

			expiration, err := getFloat(option, "expiration")
			if err != nil {
				return nil, err
			}
			t := time.Unix(int64(expiration), 0)
			contract.Expiration = t.UTC().Format("20060102")

			lastTradeDate, err := getFloat(option, "lastTradeDate")
			if err != nil {
				return nil, err
			}
			contract.LastTradeDate = time.Unix(int64(lastTradeDate), 0)

			contracts = append(contracts, contract)
		}
	}

	return contracts, nil
}

// parsePrice extracts the security price from the OCS.
func parsePrice(ocs map[string]interface{}) (float64, error) {
	quote, err := get(ocs, "quote")
	if err != nil {
		return 0, err
	}
	if quote == nil {
		return 0, fmt.Errorf("Quote was nil")
	}

	regularMarketPrice, err := get(quote, "regularMarketPrice")
	if err != nil {
		return 0, err
	}
	if regularMarketPrice == nil {
		return 0, fmt.Errorf("regularMarketPrice was nil")
	}

	price, ok := regularMarketPrice.(float64)
	if !ok {
		return 0, fmt.Errorf("regularMarketPrice was not a float64 %v", regularMarketPrice)
	}

	return price, nil
}

// parseStrikes extracts the strike prices from the OCS.
func parseStrikes(ocs map[string]interface{}) ([]float64, error) {
	strikes, err := get(ocs, "strikes")
	if err != nil {
		return nil, err
	}
	if strikes == nil {
		return nil, fmt.Errorf("Strikes was nil")
	}

	var strikePrices []float64
	for _, val := range strikes.([]interface{}) {
		if val == nil {
			return nil, fmt.Errorf("Strike value was nil")
		}
		strikePrices = append(strikePrices, val.(float64))
	}

	return strikePrices, nil
}

// parseOCS extracts all of the interesting information from the raw Yahoo! format.
func parseOCS(m map[string]interface{}, security sec.Security) (sec.Security, error) {
	optionChain, err := get(m, "optionChain")
	if err != nil {
		return security, err
	}

	// If the option chain has an error, stop processing.
	_, err = get(optionChain, "error")
	if err != nil {
		return security, err
	}

	result, err := get(optionChain, "result")
	if err != nil {
		return security, err
	}

	if len(result.([]interface{})) != 1 {
		return security, fmt.Errorf("len(result) != 1 %v", m)
	}

	ocs := result.([]interface{})[0].(map[string]interface{})

	hasMiniOptions, err := get(ocs, "hasMiniOptions")
	if err != nil {
		return security, err
	}
	hasMO, ok := hasMiniOptions.(bool)
	if ok {
		security.HasMiniOptions = hasMO
	}

	security.Puts, err = parseContracts(ocs, "puts")
	if err != nil {
		return security, err
	}

	security.Calls, err = parseContracts(ocs, "calls")
	if err != nil {
		return security, err
	}

	security.Price, err = parsePrice(ocs)
	if err != nil {
		return security, err
	}

	security.Strikes, err = parseStrikes(ocs)
	if err != nil {
		return security, err
	}

	return security, nil
}

// Symbol looks up a single ticker symbol on Yahoo! Finance and returns the associated JSON data block.
func Symbol(security sec.Security) (sec.Security, error) {
	url := "https://query1.finance.yahoo.com/v7/finance/options/" + security.Ticker

	// If the data is in the cache we can exit early.
	ocs, err := cache.Read(url)
	if err == nil {
		security, err = parseOCS(ocs, security)
		if err == nil {
			return security, nil
		}
	}

	response, err := web.Request2(url, map[string]string{})
	if err != nil {
		return security, err
	}
	if response.StatusCode != 200 {
		return security, fmt.Errorf("Unexpected response code %d getting symbol '%s'", response.StatusCode, security.Ticker)
	}

	s, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return security, err
	}

	dec := json.NewDecoder(strings.NewReader(string(s)))
	var i interface{}
	err = dec.Decode(&i)
	if err != nil {
		return security, err
	}

	// If the web request was successful we should get back a
	// map in JSON form. If it failed we should get back an error
	// message in string form. Make sure we got a map.
	ocs, ok := i.(map[string]interface{})
	if !ok {
		return security, fmt.Errorf("RequestJSON: Expected a map, got: /%s/", string(s))
	}

	security, err = parseOCS(ocs, security)
	if err != nil {
		return security, fmt.Errorf("Error parsing OCS %s", err)
	}

	// Only update the cache if the options fields were populated.
	// Sometimes Yahoo returns empty sets when queried.
	if security.HasOptions() {
		cache.Update(url, ocs)
	}

	return security, nil
}
