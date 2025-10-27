//go:generate ../../../tools/readme_config_includer/generator
package arc

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tinylib/msgp/msgp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
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
	httpconfig.HTTPClientConfig
	Log telegraf.Logger `toml:"-"`

	client *http.Client
}

type arcColumnarData struct {
	Measurement string                 `msgpack:"m"`
	Columns     map[string]interface{} `msgpack:"columns"`
}

func (*Arc) SampleConfig() string {
	return sampleConfig
}

func (a *Arc) Init() error {
	if a.URL == "" {
		a.URL = "http://localhost:8000/api/v1/write/msgpack"
	}

	if a.ContentEncoding == "" {
		a.ContentEncoding = "gzip"
	}

	if a.Timeout == 0 {
		a.Timeout = config.Duration(5 * time.Second)
	}

	return nil
}

func (a *Arc) Connect() error {
	ctx := context.Background()
	client, err := a.HTTPClientConfig.CreateClient(ctx, a.Log)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	a.client = client

	return nil
}

func (a *Arc) Close() error {
	if a.client != nil {
		a.client.CloseIdleConnections()
	}
	return nil
}

func (a *Arc) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	groups := make(map[string]*group)
	for _, m := range metrics {
		name := m.Name()
		if _, found := groups[name]; !found {
			numCols := len(m.FieldList()) + len(m.TagList()) + 1
			groups[name] = &group{name: name, columns: make(map[string][]interface{}, numCols)}
		}
		groups[name].add(m)
	}

	messages := make([]*arcColumnarData, 0, len(groups))
	for _, g := range groups {
		msg, err := g.produceMessage()
		if err != nil {
			a.Log.Error(err)
			continue
		}
		messages = append(messages, msg)
	}

	var data interface{}
	switch len(messages) {
	case 0:
		return nil
	case 1:
		data = map[string]interface{}{
			"m":       messages[0].Measurement,
			"columns": messages[0].Columns,
		}
	default:
		dataArray := make([]interface{}, len(messages))
		for i, msg := range messages {
			dataArray[i] = map[string]interface{}{
				"m":       msg.Measurement,
				"columns": msg.Columns,
			}
		}
		data = dataArray
	}

	var buf bytes.Buffer
	writer := msgp.NewWriter(&buf)
	err := writer.WriteIntf(data)
	if err != nil {
		return fmt.Errorf("marshalling message failed: %w", err)
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush MessagePack writer: %w", err)
	}

	payload := buf.Bytes()

	if a.ContentEncoding == "gzip" {
		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		if _, err := gzipWriter.Write(payload); err != nil {
			return fmt.Errorf("failed to gzip payload: %w", err)
		}
		if err := gzipWriter.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
		payload = buf.Bytes()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", a.URL, bytes.NewReader(payload))
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
			ContentEncoding: "gzip",
		}
	})
}
