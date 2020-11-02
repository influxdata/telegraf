package avro

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type SchemaRegistry struct {
	url string
}

const (
	schemaByID = "%s/schemas/ids/%d"
	timeout    = 2 * time.Second
)

func NewSchemaRegistry(url string) *SchemaRegistry {
	return &SchemaRegistry{url: url}
}

func (sr *SchemaRegistry) getSchema(id int) (string, error) {
	resp, err := http.Get(fmt.Sprintf(schemaByID, sr.url, id))
	if nil != err {
		log.Printf("E! SchemaRegistry: %s", err)
		return "", err
	}

	var jsonResponse map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	schema, ok := jsonResponse["schema"]
	if !ok {
		log.Printf("E! SchemaRegistry: malformed response!")
		return "", fmt.Errorf("malformed respose from schema registry")
	}

	schemaValue, ok := schema.(string)
	if !ok {
		log.Printf("E! SchemaRegistry: schema %s is not a string", schema)
		return "", fmt.Errorf("malformed respose from schema registry")
	}

	return schemaValue, nil
}
