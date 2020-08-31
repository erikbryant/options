package csv

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// GetFile returns the contents of a CSV file, minus the header and any blank lines.
func GetFile(file string) ([]string, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s", file)
	}

	lines := strings.Split(string(contents), "\n")

	// Strip trailing blank lines
	for lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Skip the header line
	return lines[1:], nil
}
