package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf/config"
)

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

	tlsCfg, err := p.HTTPClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   tlsCfg,
				DisableKeepAlives: true,
			},
		}
		defer client.CloseIdleConnections()
		if err := p.refreshHTTPServices(p.HTTPSDConfig.URL, client); err != nil {
			p.Log.Errorf("Unable to refresh HTTP scraped services: %v", err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(queryInterval):
				err := p.refreshHTTPServices(p.HTTPSDConfig.URL, client)
				if err != nil {
					p.Log.Errorf("Unable to refresh HTTP scraped services: %v", err)
				}
			}
		}
	}()

	return nil
}

func (p *Prometheus) refreshHTTPServices(sdURL string, client *http.Client) error {
	services := make(map[string]urlAndAddress)
	req, err := http.NewRequest("GET", sdURL, nil)
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery failed with status %q", resp.Status)
	}

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body failed: %w", err)
	}

	var result []httpSDOutput
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unmarshalling JSON failed: %w", err)
	}

	// Validate the response
	if len(result) == 0 {
		p.Log.Warnf("Service discovery returned no results")
	}

	for _, sdOutputItem := range result {
		for _, targetValue := range sdOutputItem.Targets {
			if !strings.HasPrefix(targetValue, "http://") && !strings.HasPrefix(targetValue, "https://") {
				targetValue = "http://" + targetValue
			}

			targetURL, err := url.Parse(targetValue)
			if err != nil {
				p.Log.Warnf("Failed to parse target %q", targetValue)
				continue
			}
			service := urlAndAddress{
				url:         targetURL,
				originalURL: targetURL,
				// in this case target labels should just be added to the tags
				tags: sdOutputItem.Labels,
			}
			services[service.url.String()] = service
		}
	}

	p.lock.Lock()
	p.httpServices = services
	p.lock.Unlock()

	return nil
}
