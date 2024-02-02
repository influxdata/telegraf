package avro

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/linkedin/goavro/v2"
)

type schemaAndCodec struct {
	Schema string
	Codec  *goavro.Codec
}

type schemaRegistry struct {
	url      string
	username string
	password string
	cache    map[int]*schemaAndCodec
	client   *http.Client
}

const schemaByID = "%s/schemas/ids/%d"

func newSchemaRegistry(addr string, caCertPath string) (*schemaRegistry, error) {
	var client *http.Client
	var tlsCfg *tls.Config
	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsCfg = &tls.Config{
			RootCAs: caCertPool,
		}
	}
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			MaxIdleConns:    10,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("parsing registry URL failed: %w", err)
	}

	var username, password string
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	registry := &schemaRegistry{
		url:      u.String(),
		username: username,
		password: password,
		cache:    make(map[int]*schemaAndCodec),
		client:   client,
	}

	return registry, nil
}

func (sr *schemaRegistry) getSchemaAndCodec(id int) (*schemaAndCodec, error) {
	if v, ok := sr.cache[id]; ok {
		return v, nil
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(schemaByID, sr.url, id), nil)
	if err != nil {
		return nil, err
	}

	if sr.username != "" {
		req.SetBasicAuth(sr.username, sr.password)
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
		return nil, fmt.Errorf("malformed response from schema registry: no 'schema' key")
	}

	schemaValue, ok := schema.(string)
	if !ok {
		return nil, fmt.Errorf("malformed response from schema registry: %v cannot be cast to string", schema)
	}
	codec, err := goavro.NewCodec(schemaValue)
	if err != nil {
		return nil, err
	}
	retval := &schemaAndCodec{Schema: schemaValue, Codec: codec}
	sr.cache[id] = retval
	return retval, nil
}
