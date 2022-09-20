package avro

import (
	"encoding/json"
	"fmt"
	"github.com/linkedin/goavro/v2"
	"net/http"
)

type SchemaAndCodec struct {
	Schema string
	Codec  *goavro.Codec
}

type SchemaRegistry struct {
	url   string
	cache map[int]*SchemaAndCodec
}

const (
	schemaByID = "%s/schemas/ids/%d"
)

func NewSchemaRegistry(url string) *SchemaRegistry {
	return &SchemaRegistry{url: url, cache: make(map[int]*SchemaAndCodec)}
}

func (sr *SchemaRegistry) getSchemaAndCodec(id int) (*SchemaAndCodec, error) {
	if v, ok := sr.cache[id]; ok {
		return v, nil
	}
	resp, err := http.Get(fmt.Sprintf(schemaByID, sr.url, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jsonResponse map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return nil, err
	}

	schema, ok := jsonResponse["schema"]
	if !ok {
		return nil, fmt.Errorf("malformed respose from schema registry: no 'schema' key")
	}

	schemaValue, ok := schema.(string)
	if !ok {
		return nil, fmt.Errorf("malformed respose from schema registry: %v cannot be cast to string", schema)
	}
	codec, err := goavro.NewCodec(schemaValue)
	if err != nil {
		return nil, err
	}
	retval := &SchemaAndCodec{Schema: schemaValue, Codec: codec}
	sr.cache[id] = retval
	return retval, nil
}
