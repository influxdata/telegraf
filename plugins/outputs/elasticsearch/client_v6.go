package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	estransport6 "github.com/elastic/go-elasticsearch/v6/estransport"

	"github.com/influxdata/telegraf"
)

type clientV6 struct {
	client          *elasticsearch6.Client
	httpClient      *http.Client
	log             telegraf.Logger
	cancelDiscovery context.CancelFunc
	discoveryWG     sync.WaitGroup
}

func newClientV6(transport roundTripper) (client, error) {
	c, err := elasticsearch6.NewClient(elasticsearch6.Config{
		Addresses: transport.urls,
		Transport: transport,
		// Telegraf owns write retries; replaying a bulk request can duplicate documents.
		DisableRetry:       true,
		ConnectionPoolFunc: preserveConnectionPool(estransport6.NewConnectionPool),
	})
	if err != nil {
		return nil, err
	}
	client := &clientV6{client: c, httpClient: transport.httpClient, log: transport.log}
	if transport.enableSniffer {
		// The v6 client exposes only DiscoverNodes(), so in-flight calls cannot be canceled.
		ctx, cancel := context.WithCancel(context.Background())
		client.cancelDiscovery = cancel
		client.discoveryWG.Go(func() {
			startDiscovery(ctx, transport.discoveryInterval, c.DiscoverNodes, transport.log)
		})
	}
	return client, nil
}

func (c *clientV6) close() {
	if c.cancelDiscovery != nil {
		c.cancelDiscovery()
		c.discoveryWG.Wait()
	}
	c.httpClient.CloseIdleConnections()
}

func (c *clientV6) bulk(ctx context.Context, documents []document) error {
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
