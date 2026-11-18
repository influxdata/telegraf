package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	elasticsearch5 "github.com/elastic/go-elasticsearch/v5"
	esapi5 "github.com/elastic/go-elasticsearch/v5/esapi"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type clientV5 struct {
	client  *elasticsearch5.Client
	log     telegraf.Logger
	encoder *internal.GzipEncoder
	headers http.Header
}

func (a *Elasticsearch) newClientV5() (client, error) {
	transport, err := a.newRoundTripper()
	if err != nil {
		return nil, err
	}
	c, err := elasticsearch5.NewClient(elasticsearch5.Config{
		Addresses: a.URLs,
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}
	client := &clientV5{client: c, log: a.Log, headers: a.clientHeaders()}
	if a.EnableGzip {
		encoder, err := internal.NewGzipEncoder()
		if err != nil {
			return nil, fmt.Errorf("setting up gzip encoder failed: %w", err)
		}
		client.encoder = encoder
	}
	return client, nil
}

func (*clientV5) close() {}

func (c *clientV5) compressRequestBodyIfEnabled(body []byte) (*bytes.Reader, map[string]string, error) {
	if c.encoder == nil {
		return bytes.NewReader(body), nil, nil
	}
	data, err := c.encoder.Encode(body)
	if err != nil {
		return nil, nil, err
	}
	return bytes.NewReader(data), map[string]string{"Content-Encoding": "gzip"}, nil
}

func (c *clientV5) requestHeaders(additional map[string]string) http.Header {
	headers := make(http.Header, len(c.headers)+len(additional))
	for name, value := range additional {
		headers.Add(name, value)
	}
	for name, values := range c.headers {
		for _, value := range values {
			headers.Add(name, value)
		}
	}
	return headers
}

func (c *clientV5) bulk(ctx context.Context, documents []document) error {
	body, err := encodeBulk(documents, "metrics")
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	requestBody, headers, err := c.compressRequestBodyIfEnabled(body.Bytes())
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	request := esapi5.BulkRequest{Body: requestBody, Header: c.requestHeaders(headers)}
	res, err := request.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	defer res.Body.Close()
	return checkBulkResponse(res.StatusCode, res.Body, c.log)
}

func (c *clientV5) templateExists(ctx context.Context, name string) (bool, error) {
	request := esapi5.IndicesExistsTemplateRequest{Name: []string{name}, Header: c.requestHeaders(nil)}
	res, err := request.Do(ctx, c.client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if err := checkForError(res.StatusCode, res.Body); err != nil {
		return false, err
	}
	return true, nil
}

func (c *clientV5) putTemplate(ctx context.Context, name, body string) error {
	requestBody, headers, err := c.compressRequestBodyIfEnabled([]byte(body))
	if err != nil {
		return err
	}
	request := esapi5.IndicesPutTemplateRequest{Name: name, Body: requestBody, Header: c.requestHeaders(headers)}
	res, err := request.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err := checkForError(res.StatusCode, res.Body); err != nil {
		return err
	}
	var response templateResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("decoding message body failed: %w", err)
	}
	return nil
}
