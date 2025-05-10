package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf/config"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HTTPSDConfig struct {
	Enabled       bool            `toml:"enabled"`
	URL           string          `toml:"url"`
	QueryInterval config.Duration `toml:"query_interval"`
}

// standard output for http service discovery is here https://prometheus.io/docs/prometheus/latest/http_sd/#http_sd-format
type httpSDOutput struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func (p *Prometheus) startHTTPSD(ctx context.Context) error {
	// default settings
	var queryInterval time.Duration
	if p.HTTPSDConfig.QueryInterval == 0 {
		queryInterval = time.Duration(5) * time.Minute
	} else {
		queryInterval = time.Duration(p.HTTPSDConfig.QueryInterval)
	}

	if p.HTTPSDConfig.URL == "" {
		p.HTTPSDConfig.URL = "http://localhost:9000/service-discovery"
	}

	httpSDUrl, err := url.Parse(p.HTTPSDConfig.URL)
	if err != nil {
		return fmt.Errorf("cannot parse the http service discovery url: %w", err)
	}

	tlsCfg, err := p.HTTPClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	p.wg.Add(1)
	go func() {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   tlsCfg,
				DisableKeepAlives: true,
			},
		}
		defer client.CloseIdleConnections()
		defer p.wg.Done()
		err := p.refreshHTTPServices(httpSDUrl, client)
		if err != nil {
			p.Log.Errorf("Unable to refresh HTTP scraped services: %v", err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(queryInterval):
				err := p.refreshHTTPServices(httpSDUrl, client)
				if err != nil {
					p.Log.Errorf("Unable to refresh HTTP scraped services: %v", err)
				}
			}
		}
	}()

	return nil
}

func (p *Prometheus) refreshHTTPServices(sdURL *url.URL, client HTTPClient) error {
	refreshHTTPServices := make(map[string]urlAndAddress)

	p.Log.Debugf("Refreshing HTTP services")
	req, err := http.NewRequest("GET", sdURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http service discovery url %q returned status %q", sdURL.String(), resp.Status)
	}

	var body []byte
	if p.ContentLengthLimit != 0 {
		limit := int64(p.ContentLengthLimit)
		lr := io.LimitReader(resp.Body, limit)
		body, err = io.ReadAll(lr)
		if err != nil {
			return err
		}
	} else {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	var result []httpSDOutput
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unmarshalling JSON failed: %w", err)
	}

	for _, sdOutputItem := range sdOutput {
		for _, targetValue := range sdOutputItem.Targets {
			// http service discovery returns <host>:<port> pairs so we default to appending http to all hosts
			targetValue = "http://" + targetValue

			targetURL, err := url.Parse(targetValue)
			if err != nil {
				p.Log.Warnf("Failed to parse target %q", targetValue)
				break
			}
			service := urlAndAddress{
				url:         targetURL,
				originalURL: targetURL,
				// in this case target labels should just be added to the tags
				tags: sdOutputItem.Labels,
			}
			refreshHTTPServices[service.url.String()] = service
		}
	}

	p.lock.Lock()
	p.httpServices = refreshHTTPServices
	p.lock.Unlock()

	return nil
}
