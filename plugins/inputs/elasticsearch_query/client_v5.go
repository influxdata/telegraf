package elasticsearch_query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	elasticsearch5 "github.com/elastic/go-elasticsearch/v5"

	"github.com/influxdata/telegraf"
)

type clientV5 struct {
	client     *elasticsearch5.Client
	httpClient *http.Client
	log        telegraf.Logger
}

func newClientV5(cfg clientConfig) (client, error) {
	c, err := elasticsearch5.NewClient(elasticsearch5.Config{
		Addresses: cfg.urls,
		Username:  cfg.username,
		Password:  cfg.password,
		Transport: roundTripper{client: cfg.httpClient},
	})
	if err != nil {
		cfg.httpClient.CloseIdleConnections()
		return nil, fmt.Errorf("creating ElasticSearch client failed: %w", err)
	}

	return &clientV5{client: c, httpClient: cfg.httpClient, log: cfg.log}, nil
}

func (c *clientV5) close() {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
}

func (c *clientV5) getFieldMapping(ctx context.Context, index, field string) (map[string]interface{}, error) {
	res, err := c.client.Indices.GetFieldMapping(
		[]string{field},
		c.client.Indices.GetFieldMapping.WithContext(ctx),
		c.client.Indices.GetFieldMapping.WithIndex(index),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := checkForError(res.StatusCode, res.Body); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding message body failed: %w", err)
	}
	return result, nil
}

func (c *clientV5) query(ctx context.Context, aggregation *aggregation) (interface{}, int64, error) {
	data, err := aggregation.buildSearchBody(c.log)
	if err != nil {
		return nil, 0, err
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(aggregation.Index),
		c.client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if err := checkForError(res.StatusCode, res.Body); err != nil {
		return nil, 0, err
	}

	var result searchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("decoding message body failed: %w", err)
	}
	if len(result.Aggregations) == 0 {
		return nil, result.totalHits(), nil
	}
	return result.Aggregations, result.totalHits(), nil
}
