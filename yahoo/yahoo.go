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
	val, err := get(i, key)
	if err != nil {
		return -1, err
	}
	if val == nil {
		return -1, fmt.Errorf("%s was nil", key)
	}

	raw, err := get(val, "raw")
	if err != nil {
		return -1, err
	}
	if raw == nil {
		return -1, fmt.Errorf("%s[\"raw\"] was nil", key)
	}

	return raw.(float64), nil
}

// getFmtString safely extracts val[key]["fmt"] as a string.
func getFmtString(i interface{}, key string) (string, error) {
	val, err := get(i, key)
	if err != nil {
		return "", err
	}
	if val == nil {
		return "", fmt.Errorf("%s was nil", key)
	}

	f, err := get(val, "fmt")
	if err != nil {
		return "", err
	}
	if f == nil {
		return "", fmt.Errorf("%s[\"fmt\"] was nil", key)
	}

	return f.(string), nil
}

// get reads a key from a map[string]interface{} and returns the value.
func get(i interface{}, key string) (interface{}, error) {
	if i == nil {
		return nil, fmt.Errorf("i is nil trying to get %s", key)
	}

	// TODO: verify i is map[string]interface{} before casting it.
	return i.(map[string]interface{})[key], nil
}

// parseContracts extracts the put or call (specified by 'position') options from the OCS.
func parseContracts(ocs map[string]interface{}, position string) ([]sec.Contract, error) {
	// Puts and calls.
	options, err := get(ocs, "contracts")
	if err != nil {
		return nil, err
	}
	if options == nil {
		return nil, fmt.Errorf("Contracts was nil")
	}

	optionsObject, err := get(options, position)
	if err != nil {
		return nil, err
	}

	if optionsObject == nil {
		return nil, fmt.Errorf("Nil value for contracts %s", position)
	}

	var contracts []sec.Contract

	for _, option := range optionsObject.([]interface{}) {
		var contract sec.Contract
		var err error

		contract.Strike, err = getRawFloat(option, "strike")
		if err != nil {
			return nil, err
		}

		contract.Last, err = getRawFloat(option, "lastPrice")
		if err != nil {
			return nil, err
		}

		contract.Bid, err = getRawFloat(option, "bid")
		if err != nil {
			return nil, err
		}

		contract.Ask, err = getRawFloat(option, "ask")
		if err != nil {
			return nil, err
		}

		contract.Expiration, err = getFmtString(option, "expiration")
		if err != nil {
			return nil, err
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// parsePrice extracts the security price from the OCS.
func parsePrice(ocs map[string]interface{}) (float64, error) {
	meta, err := get(ocs, "meta")
	if err != nil {
		return 0, err
	}
	if meta == nil {
		return 0, fmt.Errorf("Meta was nil")
	}

	quote, err := get(meta, "quote")
	if err != nil {
		return 0, err
	}
	if quote == nil {
		return 0, fmt.Errorf("Quote was nil")
	}

	tmp, err := get(quote, "regularMarketPrice")
	if err != nil {
		return 0, err
	}
	return tmp.(float64), nil
}

// parseStrikes extracts the strike prices from the OCS.
func parseStrikes(ocs map[string]interface{}) ([]float64, error) {
	meta, err := get(ocs, "meta")
	if err != nil {
		return nil, err
	}

	strikes, err := get(meta, "strikes")
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
func parseOCS(ocs map[string]interface{}, security sec.Security) (sec.Security, error) {
	var err error

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

const headerToken = "root.App.main = "
const footerToken = `;\n}\(this\)\);`

// extractJSON extracts the JSON block from the received HTML.
func extractJSON(response string) (map[string]interface{}, error) {
	// Isolate the JSON from the surrounding HTML.
	var re = regexp.MustCompile(headerToken)
	removeHeader := re.Split(response, 2)
	if len(removeHeader) != 2 {
		return nil, fmt.Errorf("Failed to find header token %s", response)
	}

	re = regexp.MustCompile(footerToken)
	removeFooter := re.Split(removeHeader[1], 2)
	if len(removeFooter) != 2 {
		return nil, fmt.Errorf("Failed to find footer token %s", response)
	}
	jsonString := removeFooter[0]

	if jsonString[0] != '{' {
		return nil, fmt.Errorf("JSON string is missing the '{' %s", response)
	}

	if jsonString[len(jsonString)-1] != '}' {
		return nil, fmt.Errorf("JSON string is missing the '}' %s", response)
	}

	// Convert the string form to JSON object form.
	var m interface{}
	dec := json.NewDecoder(strings.NewReader(string(jsonString)))
	err := dec.Decode(&m)
	if err != nil {
		return nil, err
	}

	// If the parsing was successful we should get back a
	// map in JSON form. Make sure we got a map.
	jsonObject, ok := m.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("RequestJSON: Expected a map, got: /%s/", string(jsonString[0]))
	}

	return jsonObject, nil
}

// extractOCS finds the OptionContractsStore and returns it.
func extractOCS(jsonObject map[string]interface{}) (map[string]interface{}, error) {
	if jsonObject == nil {
		return nil, fmt.Errorf("jsonObject is nil")
	}

	context := jsonObject["context"]
	if context == nil {
		return nil, fmt.Errorf("context is nil")
	}

	dispatcher := context.(map[string]interface{})["dispatcher"]
	if dispatcher == nil {
		return nil, fmt.Errorf("dispatcher is nil")
	}
	stores := dispatcher.(map[string]interface{})["stores"]
	if stores == nil {
		return nil, fmt.Errorf("stores is nil")
	}

	optionContractsStore := stores.(map[string]interface{})["OptionContractsStore"]
	if optionContractsStore == nil {
		return nil, fmt.Errorf("OptionContractsStore is nil")
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

	security, err = parseOCS(ocs, security)
	if err != nil {
		return security, fmt.Errorf("Error parsing OCS %s", err)
	}

	return security, nil
}
