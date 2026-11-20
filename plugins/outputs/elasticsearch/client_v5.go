package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v5"
	"github.com/elastic/go-elasticsearch/v5/esapi"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type clientV5 struct {
	client      *elasticsearch.Client
	log         telegraf.Logger
	encoder     *internal.GzipEncoder
	headers     http.Header
	bodyHeaders http.Header
}

func (a *Elasticsearch) newClientV5() (client, error) {
	transport, err := a.newRoundTripper()
	if err != nil {
		return nil, err
	}
	c, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: a.URLs,
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}
	headers := a.clientHeaders()
	client := &clientV5{client: c, log: a.Log, headers: headers, bodyHeaders: headers}
	if a.EnableGzip {
		// Content-Encoding applies only to requests with a compressed body.
		client.bodyHeaders = headers.Clone()
		if client.bodyHeaders == nil {
			client.bodyHeaders = make(http.Header)
		}
		client.bodyHeaders.Set("Content-Encoding", "gzip")
		encoder, err := internal.NewGzipEncoder()
		if err != nil {
			return nil, fmt.Errorf("setting up gzip encoder failed: %w", err)
		}
		client.encoder = encoder
	}
	return client, nil
}

func (*clientV5) close() {}

func (c *clientV5) bulk(ctx context.Context, documents []document) error {
	body, err := encodeBulk(documents, "metrics")
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	requestBody := bytes.NewReader(body.Bytes())
	if c.encoder != nil {
		data, err := c.encoder.Encode(body.Bytes())
		if err != nil {
			return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
		}
		requestBody = bytes.NewReader(data)
	}
	request := esapi.BulkRequest{Body: requestBody, Header: c.bodyHeaders.Clone()}
	res, err := request.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	defer res.Body.Close()
	return checkBulkResponse(res.StatusCode, res.Body, c.log)
}

func (c *clientV5) templateExists(ctx context.Context, name string) (bool, error) {
	request := esapi.IndicesExistsTemplateRequest{Name: []string{name}, Header: c.headers.Clone()}
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
	requestBody := bytes.NewReader([]byte(body))
	if c.encoder != nil {
		data, err := c.encoder.Encode([]byte(body))
		if err != nil {
			return err
		}
		requestBody = bytes.NewReader(data)
	}
	request := esapi.IndicesPutTemplateRequest{Name: name, Body: requestBody, Header: c.bodyHeaders.Clone()}
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
