package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

const cacheDir = "./yahoo-cache/"

func sanitize(id string) string {
	id = strings.ReplaceAll(id, "/", "-")
	return id
}

// Read reads an object from the cache (if it is present)
func Read(id string) (map[string]interface{}, error) {
	object := path.Join(cacheDir, sanitize(id))

	contents, err := ioutil.ReadFile(object)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s %s", object, err)
	}

	// Convert the string form to JSON object form.
	var m interface{}
	dec := json.NewDecoder(strings.NewReader(string(contents)))
	err = dec.Decode(&m)
	if err != nil {
		return nil, err
	}

	// If the parsing was successful we should get back a
	// map in JSON form. Make sure we got a map.
	jsonObject, ok := m.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("RequestJSON: Expected a map, got: /%s/", contents)
	}

	return jsonObject, nil
}

// Update inserts an object into the cache
func Update(id string, contents map[string]interface{}) {
	object := path.Join(cacheDir, sanitize(id))

	s, err := json.MarshalIndent(contents, "", " ")
	if err != nil {
		return
	}

	ioutil.WriteFile(object, s, 0644)
}
