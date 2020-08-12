package yahoo

import (
	"encoding/json"
	"fmt"
	sec "github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
	"regexp"
	"strings"
)

// getRawFloat safely extracts val[key]["raw"] as a float64.
func getRawFloat(i interface{}, key string) (float64, error) {
	val := get(i, key)

	if val == nil {
		return -1, fmt.Errorf("%s was nil", key)
	}

	raw := get(val, "raw")

	if raw == nil {
		return -1, fmt.Errorf("%s[\"raw\"] was nil", key)
	}

	return raw.(float64), nil
}

// getFmtString safely extracts val[key]["fmt"] as a string.
func getFmtString(i interface{}, key string) (string, error) {
	val := get(i, key)

	if val == nil {
		return "", fmt.Errorf("%s was nil", key)
	}

	f := get(val, "fmt")

	if f == nil {
		return "", fmt.Errorf("%s[\"fmt\"] was nil", key)
	}

	return f.(string), nil
}

// get reads a key from a map[string]interface{} and returns it.
func get(i interface{}, key string) interface{} {
	if i == nil {
		fmt.Println("i is nil trying to get", key)
		return nil
	}

	return i.(map[string]interface{})[key]
}

// ParseOCS extracts all of the interesting information from the raw Yahoo! format.
func ParseOCS(ocs map[string]interface{}, security sec.Security) (sec.Security, error) {
	// The price of the underlying security.
	meta := get(ocs, "meta")
	if meta == nil {
		return security, fmt.Errorf("Meta was nil")
	}

	quote := get(meta, "quote")
	if quote == nil {
		return security, fmt.Errorf("Quote was nil")
	}

	security.Price = get(quote, "regularMarketPrice").(float64)

	// The list of strike prices. Arrays cannot be typecast. Make a copy instead.
	strikes := get(meta, "strikes")
	if strikes == nil {
		return security, fmt.Errorf("Strikes was nil")
	}
	for _, val := range strikes.([]interface{}) {
		if val == nil {
			return security, fmt.Errorf("Val was nil")
		}
		security.Strikes = append(security.Strikes, val.(float64))
	}

	contracts := get(ocs, "contracts")
	if contracts == nil {
		return security, fmt.Errorf("Contracts was nil")
	}

	// The puts.
	puts := get(contracts, "puts")
	if puts == nil {
		return security, fmt.Errorf("Puts was nil")
	}

	for _, val := range puts.([]interface{}) {
		var put sec.Contract
		var err error

		put.Strike, err = getRawFloat(val, "strike")
		if err != nil {
			return security, err
		}

		put.Last, err = getRawFloat(val, "lastPrice")
		if err != nil {
			return security, err
		}

		put.Bid, err = getRawFloat(val, "bid")
		if err != nil {
			return security, err
		}

		put.Ask, err = getRawFloat(val, "ask")
		if err != nil {
			return security, err
		}

		put.Expiration, err = getFmtString(val, "expiration")
		if err != nil {
			return security, err
		}

		security.Puts = append(security.Puts, put)
	}

	return security, nil
}

// extract extracts the JSON block from the received HTML.
func extractJSON(response string) (map[string]interface{}, error) {
	var re = regexp.MustCompile("root.App.main = ")
	json1 := re.Split(response, 2)
	re = regexp.MustCompile(`;\n}\(this\)\);`)
	json2 := re.Split(json1[1], 2)

	dec := json.NewDecoder(strings.NewReader(string(json2[0])))
	var m interface{}
	err := dec.Decode(&m)
	if err != nil {
		return nil, err
	}

	// If the web request was successful we should get back a
	// map in JSON form. If it failed we should get back an error
	// message in string form. Make sure we got a map.
	f, ok := m.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("RequestJSON: Expected a map, got: /%s/", string(json2[0]))
	}

	return f, nil
}

func extractOCS(f map[string]interface{}) (map[string]interface{}, error) {
	context := f["context"]
	dispatcher := context.(map[string]interface{})["dispatcher"]
	stores := dispatcher.(map[string]interface{})["stores"]
	optionContractsStore := stores.(map[string]interface{})["OptionContractsStore"]

	if optionContractsStore == nil {
		return nil, fmt.Errorf("OptionsContractStore is nil")
	}

	return optionContractsStore.(map[string]interface{}), nil
}

// Symbol looks up a single ticker symbol on Yahoo! Finance and returns the associated JSON data block.
func Symbol(security sec.Security) (sec.Security, error) {
	url := "https://finance.yahoo.com/quote/" + security.Ticker + "/options?p=" + security.Ticker

	response, err := web.Request(url, map[string]string{})
	if err != nil {
		return security, err
	}

	f, err := extractJSON(response)
	if err != nil {
		return security, err
	}

	ocs, err := extractOCS(f)
	if err != nil {
		return security, err
	}

	security, err = ParseOCS(ocs, security)
	if err != nil {
		return security, fmt.Errorf("Error parsing OCS %s", err)
	}

	return security, nil
}
