package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	estransport6 "github.com/elastic/go-elasticsearch/v6/estransport"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type clientV6 struct {
	client          *elasticsearch6.Client
	log             telegraf.Logger
	encoder         *internal.GzipEncoder
	cancelDiscovery context.CancelFunc
	discoveryWG     sync.WaitGroup
}

func (a *Elasticsearch) newClientV6() (client, error) {
	transport, err := a.newRoundTripper()
	if err != nil {
		return nil, err
	}
	c, err := elasticsearch6.NewClient(elasticsearch6.Config{
		Addresses: a.URLs,
		Header:    a.clientHeaders(),
		Transport: transport,
		// Telegraf owns write retries; replaying a bulk request can duplicate documents.
		DisableRetry:       true,
		ConnectionPoolFunc: preserveConnectionPool(estransport6.NewConnectionPool),
	})
	if err != nil {
		return nil, err
	}
	client := &clientV6{client: c, log: a.Log}
	if a.EnableGzip {
		encoder, err := internal.NewGzipEncoder()
		if err != nil {
			return nil, fmt.Errorf("setting up gzip encoder failed: %w", err)
		}
		client.encoder = encoder
	}
	if a.EnableSniffer {
		// The discovery API accepts no context, so in-flight calls cannot be canceled.
		ctx, cancel := context.WithCancel(context.Background())
		client.cancelDiscovery = cancel
		client.discoveryWG.Go(func() {
			startDiscovery(ctx, c.DiscoverNodes, a.Log)
		})
	}
	return client, nil
}

func (c *clientV6) close() {
	if c.cancelDiscovery != nil {
		c.cancelDiscovery()
		c.discoveryWG.Wait()
	}
}

func (c *clientV6) compressRequestBodyIfEnabled(body []byte) (*bytes.Reader, map[string]string, error) {
	if c.encoder == nil {
		return bytes.NewReader(body), nil, nil
	}
	data, err := c.encoder.Encode(body)
	if err != nil {
		return nil, nil, err
	}
	return bytes.NewReader(data), map[string]string{"Content-Encoding": "gzip"}, nil
}

func (c *clientV6) bulk(ctx context.Context, documents []document) error {
	body, err := encodeBulk(documents, "metrics")
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	requestBody, headers, err := c.compressRequestBodyIfEnabled(body.Bytes())
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	res, err := c.client.Bulk(
		requestBody,
		c.client.Bulk.WithContext(ctx),
		c.client.Bulk.WithHeader(headers),
	)
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	defer res.Body.Close()
	return checkBulkResponse(res.StatusCode, res.Body, c.log)
}

func (c *clientV6) templateExists(ctx context.Context, name string) (bool, error) {
	res, err := c.client.Indices.ExistsTemplate([]string{name}, c.client.Indices.ExistsTemplate.WithContext(ctx))
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

func (c *clientV6) putTemplate(ctx context.Context, name, body string) error {
	requestBody, headers, err := c.compressRequestBodyIfEnabled([]byte(body))
	if err != nil {
		return err
	}
	res, err := c.client.Indices.PutTemplate(
		name,
		requestBody,
		c.client.Indices.PutTemplate.WithContext(ctx),
		c.client.Indices.PutTemplate.WithHeader(headers),
	)
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
