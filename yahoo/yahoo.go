package yahoo

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/web"
	"regexp"
	"strings"
)

// extract extracts the JSON block from the received HTML.
func extract() {

}

// Symbol looks up a single ticker symbol on Yahoo! Finance and returns the associated JSON data block.
func Symbol(s string) (map[string]interface{}, error) {
	url := "https://finance.yahoo.com/quote/" + s + "/options?p=" + s

	response, err := web.Request(url, map[string]string{})
	if err != nil {
		return nil, err
	}

	var re = regexp.MustCompile("root.App.main = ")
	json1 := re.Split(response, 2)
	re = regexp.MustCompile(`;\n}\(this\)\);`)
	json2 := re.Split(json1[1], 2)

	dec := json.NewDecoder(strings.NewReader(string(json2[0])))
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
		return nil, fmt.Errorf("RequestJSON: Expected a map, got: /%s/", string(json2[0]))
	}

	context := f["context"]
	dispatcher := context.(map[string]interface{})["dispatcher"]
	stores := dispatcher.(map[string]interface{})["stores"]
	optionContractsStore := stores.(map[string]interface{})["OptionContractsStore"]

	if optionContractsStore == nil {
		return nil, fmt.Errorf("OptionsContractStore is nil")
	}

	return optionContractsStore.(map[string]interface{}), nil
}
