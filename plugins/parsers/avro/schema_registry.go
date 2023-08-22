package avro

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/linkedin/goavro/v2"
)

type schemaAndCodec struct {
	Schema string
	Codec  *goavro.Codec
}

type schemaRegistry struct {
	url         string
	auth_base64 string
	cache       map[int]*schemaAndCodec
	client      http.Client
}

const schemaByID = "%s/schemas/ids/%d"

func newSchemaRegistry(url string, auth_base64 string, ca_cert_path string) (*schemaRegistry, error) {
	var client *http.Client

	caCert, err := ioutil.ReadFile(ca_cert_path)
	if err != nil {
		return nil, err
	}

	if ca_cert_path != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
				MaxIdleConns:    10,
				IdleConnTimeout: 90 * time.Second,
			},
		}
	} else {
		client = &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 90 * time.Second,
			},
		}
	}

	return &schemaRegistry{url: url, auth_base64: auth_base64, cache: make(map[int]*schemaAndCodec), client: *client}, nil
}

func (sr *schemaRegistry) getSchemaAndCodec(id int) (*schemaAndCodec, error) {
	if v, ok := sr.cache[id]; ok {
		return v, nil
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(schemaByID, sr.url, id), nil)
	if err != nil {
		return nil, err
	}

	if sr.auth_base64 != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", sr.auth_base64))
	}

	resp, err := sr.client.Do(req)
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
