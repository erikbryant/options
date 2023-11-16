package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

const cacheDir = "./web-request-cache/"

// sanitize replaces reserved filesystem characers with '-'
func sanitize(id string) string {
	id = strings.ReplaceAll(id, "/", "-")
	return id
}

// Read reads an object from the cache (if it is present)
func Read(id string) (map[string]interface{}, error) {
	object := path.Join(cacheDir, sanitize(id))

	contents, err := os.ReadFile(object)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s %s", object, err)
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal cached json %s", err)
	}

	return jsonObject, nil
}

// Update inserts an object into the cache
func Update(id string, contents map[string]interface{}) {
	object := path.Join(cacheDir, sanitize(id))

	s, err := json.MarshalIndent(contents, "", " ")
	if err != nil {
		fmt.Printf("Could not marshal contents for %v\n", contents)
		return
	}

	err = os.WriteFile(object, s, 0644)
	if err != nil {
		fmt.Println("Error writing cache file", err)
	}
}
