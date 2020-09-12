package csv

import (
	"fmt"
	"io/ioutil"
	"os"
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

// AppendFile adds to the end of a file (creates if not exist).
func AppendFile(file, contents string, truncate bool) error {
	mode := os.O_CREATE | os.O_WRONLY
	if truncate {
		mode |= os.O_TRUNC
	} else {
		mode |= os.O_APPEND
	}
	f, err := os.OpenFile(file, mode, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(contents)
	if err != nil {
		return err
	}

	return nil
}
