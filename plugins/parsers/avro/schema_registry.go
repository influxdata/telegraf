package avro

import (
	"time"
	"fmt"
	"encoding/json"
	"net/http"
)

type SchemaRegistry struct {
	url		string
	cache 	map[int]string
}

const (
	schemaByID       = "http://localhost:8081/schemas/ids/%d"
	timeout = 2 * time.Second
)

func NewSchemaRegistry(url string) *SchemaRegistry {
    return &SchemaRegistry{
    	url: url, 
    	cache: make(map[int]string),
    }
}

func (sr *SchemaRegistry) getSchema(id int) (string, error) {

	value, ok := sr.cache[id]
    if ok {
    	return value, nil
    }

	resp, err := http.Get(fmt.Sprintf(schemaByID, id))
	if nil != err {
		return "", err
	}

	fmt.Println(resp.Body)

	var schema map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&schema)

	fmt.Println(schema)

	sr.cache[id] = schema["schema"].(string)

	return schema["schema"].(string), nil
}
