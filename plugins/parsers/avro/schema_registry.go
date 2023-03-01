package avro

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/linkedin/goavro/v2"
)

type schemaAndCodec struct {
	Schema string
	Codec  *goavro.Codec
}

type schemaRegistry struct {
	url   string
	cache map[int]*schemaAndCodec
}

const schemaByID = "%s/schemas/ids/%d"

func newSchemaRegistry(url string) *schemaRegistry {
	return &schemaRegistry{url: url, cache: make(map[int]*schemaAndCodec)}
}

func (sr *schemaRegistry) getSchemaAndCodec(id int) (*schemaAndCodec, error) {
	if v, ok := sr.cache[id]; ok {
		return v, nil
	}
	resp, err := http.Get(fmt.Sprintf(schemaByID, sr.url, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
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
	retval := &schemaAndCodec{Schema: schemaValue, Codec: codec}
	sr.cache[id] = retval
	return retval, nil
}
