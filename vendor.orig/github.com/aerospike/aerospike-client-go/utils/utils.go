package utils

import (
	"encoding/base64"
	"io/ioutil"
)

// ReadFileEncodeBase64 readfs a file from disk and encodes it as base64
func ReadFileEncodeBase64(filename string) (string, error) {
	// read whole the file
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
