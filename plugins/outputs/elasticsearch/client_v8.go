package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/influxdata/telegraf"
)

type clientV8 struct {
	client *elasticsearch.BaseClient
	log    telegraf.Logger
}

func (a *Elasticsearch) newClientV8() (client, error) {
	var discoveryInterval time.Duration
	if a.EnableSniffer {
		discoveryInterval = 15 * time.Minute
	}

	c, err := elasticsearch.NewBaseClient(elasticsearch.Config{
		Addresses:             a.URLs,
		Header:                a.clientHeaders(),
		Transport:             roundTripper{httpClient: a.httpClient},
		DisableRetry:          true,
		CompressRequestBody:   a.EnableGzip,
		DiscoverNodesInterval: discoveryInterval,
		Interceptors:          []elastictransport.InterceptorFunc{a.authInterceptor()},
	})
	if err != nil {
		return nil, err
	}
	if a.EnableSniffer {
		// DiscoverNodesOnStart starts a goroutine that BaseClient.Close does not wait for.
		if err := c.DiscoverNodes(); err != nil {
			a.Log.Errorf("Discovering Elasticsearch nodes failed: %v", err)
		}
	}
	return &clientV8{client: c, log: a.Log}, nil
}

func (a *Elasticsearch) authInterceptor() elastictransport.InterceptorFunc {
	return func(next elastictransport.RoundTripFunc) elastictransport.RoundTripFunc {
		return func(req *http.Request) (*http.Response, error) {
			if err := a.setAuth(req); err != nil {
				if req.Body != nil {
					req.Body.Close()
				}
				return nil, err
			}
			return next(req)
		}
	}
}

func (c *clientV8) close() {
	if err := c.client.Close(context.Background()); err != nil {
		c.log.Errorf("Closing Elasticsearch client failed: %v", err)
	}
}

func (c *clientV8) bulk(ctx context.Context, documents []document) error {
	body, err := encodeBulk(documents, "")
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	request := esapi.BulkRequest{Body: body}
	res, err := request.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	defer res.Body.Close()
	return checkBulkResponse(res.StatusCode, res.Body, c.log)
}

func (c *clientV8) templateExists(ctx context.Context, name string) (bool, error) {
	request := esapi.IndicesExistsTemplateRequest{Name: []string{name}}
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

func (c *clientV8) putTemplate(ctx context.Context, name, body string) error {
	request := esapi.IndicesPutTemplateRequest{Name: name, Body: strings.NewReader(body)}
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
