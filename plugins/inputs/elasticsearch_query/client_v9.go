package elasticsearch_query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	// The Elasticsearch v9 client depends on the v8 transport module.
	elastictransport8 "github.com/elastic/elastic-transport-go/v8/elastictransport"
	elasticsearch9 "github.com/elastic/go-elasticsearch/v9"
	esapi9 "github.com/elastic/go-elasticsearch/v9/esapi"

	"github.com/influxdata/telegraf"
)

type clientV9 struct {
	client          *elasticsearch9.BaseClient
	httpClient      *http.Client
	log             telegraf.Logger
	cancelDiscovery context.CancelFunc
	discoveryWG     sync.WaitGroup
	closeOnce       sync.Once
}

func newClientV9(cfg clientConfig) (client, error) {
	// Use the base client to avoid retaining the full esapi API tree because this
	// plugin only uses two request types.
	c, err := elasticsearch9.NewBase(
		elasticsearch9.WithAddresses(cfg.urls...),
		elasticsearch9.WithBasicAuth(cfg.username, cfg.password),
		elasticsearch9.WithTransportOptions(
			elastictransport8.WithTransport(roundTripper{client: cfg.httpClient}),
		),
	)
	if err != nil {
		cfg.httpClient.CloseIdleConnections()
		return nil, fmt.Errorf("creating ElasticSearch client failed: %w", err)
	}

	client := &clientV9{client: c, httpClient: cfg.httpClient, log: cfg.log}
	if cfg.enableSniffer && cfg.discoveryInterval > 0 {
		// The v9 base client exposes only DiscoverNodes(), so in-flight calls
		// cannot be canceled.
		ctx, cancel := context.WithCancel(context.Background())
		client.cancelDiscovery = cancel
		client.discoveryWG.Add(1)
		go func() {
			defer client.discoveryWG.Done()
			startDiscovery(ctx, cfg.discoveryInterval, func(context.Context) error {
				return c.DiscoverNodes()
			}, cfg.log)
		}()
	}
	return client, nil
}

func (c *clientV9) close() {
	c.closeOnce.Do(func() {
		if c.cancelDiscovery != nil {
			c.cancelDiscovery()
			c.discoveryWG.Wait()
		}
		if c.client != nil {
			if err := c.client.Close(context.Background()); err != nil {
				c.log.Errorf("Closing ElasticSearch client failed: %v", err)
			}
		}
		if c.httpClient != nil {
			c.httpClient.CloseIdleConnections()
		}
	})
}

func (c *clientV9) getFieldMapping(ctx context.Context, index, field string) (map[string]interface{}, error) {
	req := esapi9.IndicesGetFieldMappingRequest{
		Index:  []string{index},
		Fields: []string{field},
	}
	res, err := req.Do(ctx, c.client)
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

func (c *clientV9) query(ctx context.Context, aggregation *aggregation) (interface{}, int64, error) {
	data, err := aggregation.buildSearchBody(c.log)
	if err != nil {
		return nil, 0, err
	}

	req := esapi9.SearchRequest{
		Index:          []string{aggregation.Index},
		Body:           bytes.NewReader(data),
		TrackTotalHits: true,
	}
	res, err := req.Do(ctx, c.client)
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
