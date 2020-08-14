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

func parseContract(c map[string]interface{}) (sec.Contract, error) {
	var contract sec.Contract
	var err error

	contract.Strike, err = getRawFloat(c, "strike")
	if err != nil {
		return contract, err
	}

	contract.Last, err = getRawFloat(c, "lastPrice")
	if err != nil {
		return contract, err
	}

	contract.Bid, err = getRawFloat(c, "bid")
	if err != nil {
		return contract, err
	}

	contract.Ask, err = getRawFloat(c, "ask")
	if err != nil {
		return contract, err
	}

	contract.Expiration, err = getFmtString(c, "expiration")
	if err != nil {
		return contract, err
	}

	return contract, nil
}

// parseOCS extracts all of the interesting information from the raw Yahoo! format.
func parseOCS(ocs map[string]interface{}, security sec.Security) (sec.Security, error) {
	// The price of the underlying security.
	meta, err := get(ocs, "meta")
	if err != nil {
		return security, err
	}
	if meta == nil {
		return security, fmt.Errorf("Meta was nil")
	}

	quote, err := get(meta, "quote")
	if err != nil {
		return security, err
	}
	if quote == nil {
		return security, fmt.Errorf("Quote was nil")
	}

	tmp, err := get(quote, "regularMarketPrice")
	if err != nil {
		return security, err
	}
	security.Price = tmp.(float64)

	// The list of strike prices. Arrays cannot be typecast. Make a copy instead.
	strikes, err := get(meta, "strikes")
	if err != nil {
		return security, err
	}
	if strikes == nil {
		return security, fmt.Errorf("Strikes was nil")
	}
	for _, val := range strikes.([]interface{}) {
		if val == nil {
			return security, fmt.Errorf("Val was nil")
		}
		security.Strikes = append(security.Strikes, val.(float64))
	}

	contracts, err := get(ocs, "contracts")
	if err != nil {
		return security, err
	}
	if contracts == nil {
		return security, fmt.Errorf("Contracts was nil")
	}

	// The puts.
	puts, err := get(contracts, "puts")
	if err != nil {
		return security, err
	}
	if puts == nil {
		return security, fmt.Errorf("Puts was nil")
	}

	for _, c := range puts.([]interface{}) {
		contract, err := parseContract(c.(map[string]interface{}))
		if err != nil {
			return security, err
		}
		security.Puts = append(security.Puts, contract)
	}

	// The calls.
	calls, err := get(contracts, "calls")
	if err != nil {
		return security, err
	}
	if calls == nil {
		return security, fmt.Errorf("Calls was nil")
	}

	for _, c := range calls.([]interface{}) {
		contract, err := parseContract(c.(map[string]interface{}))
		if err != nil {
			return security, err
		}
		security.Calls = append(security.Calls, contract)
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
