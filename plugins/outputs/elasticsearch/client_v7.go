package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	estransport7 "github.com/elastic/go-elasticsearch/v7/estransport"

	"github.com/influxdata/telegraf"
)

type clientV7 struct {
	client          *elasticsearch7.Client
	httpClient      *http.Client
	log             telegraf.Logger
	cancelDiscovery context.CancelFunc
	discoveryWG     sync.WaitGroup
}

func newClientV7(transport roundTripper) (client, error) {
	transport.bypassProductCheck = true
	c, err := elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses: transport.urls,
		Transport: transport,
		// Telegraf owns write retries; replaying a bulk request can duplicate documents.
		DisableRetry:       true,
		ConnectionPoolFunc: preserveConnectionPool(estransport7.NewConnectionPool),
	})
	if err != nil {
		return nil, err
	}
	client := &clientV7{client: c, httpClient: transport.httpClient, log: transport.log}
	if transport.enableSniffer {
		// The v7 connection-pool discovery API accepts no context, so in-flight calls cannot be canceled.
		ctx, cancel := context.WithCancel(context.Background())
		client.cancelDiscovery = cancel
		client.discoveryWG.Go(func() {
			startDiscovery(ctx, transport.discoveryInterval, c.DiscoverNodes, transport.log)
		})
	}
	return client, nil
}

func (c *clientV7) close() {
	if c.cancelDiscovery != nil {
		c.cancelDiscovery()
		c.discoveryWG.Wait()
	}
	c.httpClient.CloseIdleConnections()
}

func (c *clientV7) bulk(ctx context.Context, documents []document) error {
	body, err := encodeBulk(documents, "")
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

func (c *clientV7) templateExists(ctx context.Context, name string) (bool, error) {
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

func (c *clientV7) putTemplate(ctx context.Context, name, body string) error {
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
