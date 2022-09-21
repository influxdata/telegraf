package avro

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SchemaRegistry struct {
	url string
}

const (
	schemaByID = "%s/schemas/ids/%d"
)

func NewSchemaRegistry(url string) *SchemaRegistry {
	return &SchemaRegistry{url: url}
}

func (sr *SchemaRegistry) getSchema(id int) (string, error) {
	resp, err := http.Get(fmt.Sprintf(schemaByID, sr.url, id))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var jsonResponse map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return "", err
	}

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
