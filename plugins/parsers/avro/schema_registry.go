package avro

import (
	"encoding/json"
	"fmt"
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
	if err != nil {
		return "", err
	}

	var jsonResponse map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	schema, ok := jsonResponse["schema"]
	if !ok {
		return "", fmt.Errorf("malformed respose from schema registry: no 'schema' key")
	}

	schemaValue, ok := schema.(string)
	if !ok {
		return "", fmt.Errorf("malformed respose from schema registry: %v cannot be cast to string", schema)
	}

	return schemaValue, nil
}
