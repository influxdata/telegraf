package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// JSON ...
type JSON struct {
	Directory string
}

// Open ...
func (j *JSON) Open() error {
	fmt.Printf("JSON Open: %v\n", j.Directory)

	return os.MkdirAll(j.Directory, 0644)

}

// Close ...
func (j *JSON) Close() {
	fmt.Printf("JSON Close: %v\n", j.Directory)
}

// Load

func (j *JSON) Load(logName string) (interface{}, error) {
	fileName := filepath.Join(j.Directory, logName)
	jsonString, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	loadedState := make(map[string]interface{})
	err = json.Unmarshal(jsonString, &loadedState)
	if err != nil {
		return nil, err
	}

	return loadedState, nil
}

// Store
func (j *JSON) Store(logName string, state interface{}) error {
	fileName := filepath.Join(j.Directory, logName)

	jsonState, _ := json.Marshal(state)

	return ioutil.WriteFile(fileName, jsonState, 0644)
}

// Flush ...
func (j *JSON) Flush() {
	fmt.Printf("JSON Flush\n")
}
