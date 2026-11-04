package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	elasticsearch5 "github.com/elastic/go-elasticsearch/v5"

	"github.com/influxdata/telegraf"
)

type clientV5 struct {
	client     *elasticsearch5.Client
	httpClient *http.Client
	log        telegraf.Logger
}

func newClientV5(transport roundTripper) (client, error) {
	c, err := elasticsearch5.NewClient(elasticsearch5.Config{
		Addresses: transport.urls,
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}
	return &clientV5{client: c, httpClient: transport.httpClient, log: transport.log}, nil
}

func (c *clientV5) close() {
	c.httpClient.CloseIdleConnections()
}

func (c *clientV5) bulk(ctx context.Context, documents []document) error {
	body, err := encodeBulk(documents, "metrics")
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	res, err := c.client.Bulk(body, c.client.Bulk.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	defer res.Body.Close()
	return checkBulkResponse(res.StatusCode, res.Body, c.log)
}

func (c *clientV5) templateExists(ctx context.Context, name string) (bool, error) {
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

func (c *clientV5) putTemplate(ctx context.Context, name, body string) error {
	res, err := c.client.Indices.PutTemplate(name, strings.NewReader(body), c.client.Indices.PutTemplate.WithContext(ctx))
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
