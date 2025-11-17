//go:generate ../../../tools/readme_config_includer/generator
package arc

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tinylib/msgp/msgp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Arc struct {
	URL             string            `toml:"url"`
	APIKey          config.Secret     `toml:"api_key"`
	Database        string            `toml:"database"`
	Headers         map[string]string `toml:"headers"`
	ContentEncoding string            `toml:"content_encoding"`
	Log             telegraf.Logger   `toml:"-"`
	common_http.HTTPClientConfig

	client *http.Client
	cancel context.CancelFunc
}

func (*Arc) SampleConfig() string {
	return sampleConfig
}

func (a *Arc) Init() error {
	if a.URL == "" {
		return errors.New("url is required")
	}

	switch a.ContentEncoding {
	case "":
		a.ContentEncoding = "gzip"
	case "none", "identity", "gzip":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("unknown content encoding %q", a.ContentEncoding)
	}

	return nil
}

func (a *Arc) Connect() error {
	ctx, cancel := context.WithCancel(context.Background())
	client, err := a.HTTPClientConfig.CreateClient(ctx, a.Log)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	a.client = client
	a.cancel = cancel

	return nil
}

func (a *Arc) Close() error {
	if a.cancel != nil {
		a.cancel()
	}
	if a.client != nil {
		a.client.CloseIdleConnections()
	}
	return nil
}

func (a *Arc) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Group metrics by measurement name for columnar format
	groups := make(map[string]*group)
	for _, m := range metrics {
		name := m.Name()
		if _, found := groups[name]; !found {
			numCols := len(m.FieldList()) + len(m.TagList()) + 1
			groups[name] = &group{name: name, columns: make(map[string][]interface{}, numCols)}
		}
		groups[name].add(m)
	}

	// Extract the output messages from the groups
	messages := make([]map[string]interface{}, 0, len(groups))
	for _, g := range groups {
		msg, err := g.produceMessage()
		if err != nil {
			a.Log.Error(err)
			continue
		}
		messages = append(messages, msg)
	}

	// Prepare the data for serialization
	var data interface{}
	switch len(messages) {
	case 0:
		// If no valid message was produced, drop all metrics
		return nil
	case 1:
		// Single measurement should be sent directly as a map
		data = messages[0]
	default:
		// Multiple measurements should be sent as an array
		data = messages
	}

	var payload bytes.Buffer
	var writer io.Writer = &payload

	// Wrap with gzip writer if compression is enabled
	var gzipWriter *gzip.Writer
	if a.ContentEncoding == "gzip" {
		gzipWriter = gzip.NewWriter(&payload)
		writer = gzipWriter
	}

	// Write MessagePack data directly to the writer (gzipped or not)
	msgpWriter := msgp.NewWriter(writer)
	if err := msgpWriter.WriteIntf(data); err != nil {
		return fmt.Errorf("marshalling message failed: %w", err)
	}
	if err := msgpWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush MessagePack writer: %w", err)
	}

	// Close gzip writer before reading the buffer
	if gzipWriter != nil {
		if err := gzipWriter.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", a.URL, &payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("User-Agent", internal.ProductToken())

	if !a.APIKey.Empty() {
		apiKey, err := a.APIKey.Get()
		if err != nil {
			return fmt.Errorf("failed to get API key: %w", err)
		}
		req.Header.Set("x-api-key", apiKey.String())
		apiKey.Destroy()
	}

	if a.Database != "" {
		req.Header.Set("x-arc-database", a.Database)
	}

	if a.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	for k, v := range a.Headers {
		req.Header.Set(k, v)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to write to Arc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("arc returned status %d (failed to read response body: %w)", resp.StatusCode, err)
		}
		return fmt.Errorf("arc returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func init() {
	outputs.Add("arc", func() telegraf.Output {
		return &Arc{
			HTTPClientConfig: common_http.HTTPClientConfig{
				Timeout: config.Duration(5 * time.Second),
			},
		}
	})
}
