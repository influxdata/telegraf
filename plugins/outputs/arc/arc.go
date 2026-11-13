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

	// Reused across writes; Write is never called concurrently. The payload
	// buffer is not reused, as the HTTP transport may still read the request
	// body after the response arrived.
	gzipWriter      *gzip.Writer
	lastPayloadSize int
	nils            []byte
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

	// Run of MessagePack nil markers (0xc0) to fill absent cells of sparse columns
	a.nils = bytes.Repeat([]byte{0xc0}, 512)

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

	// Group metrics by measurement name for the columnar format, keeping the
	// order of first appearance so the payload is deterministic.
	var groups []*group
	index := make(map[string]int)
	for _, m := range metrics {
		name := m.Name()
		idx, found := index[name]
		if !found {
			idx = len(groups)
			index[name] = idx
			groups = append(groups, &group{name: name})
		}
		groups[idx].metrics = append(groups[idx].metrics, m)
	}

	// Build the column layout of each group, skipping the ones that cannot be
	// encoded. This must happen before anything is written, as the number of
	// remaining groups decides the shape of the payload.
	valid := groups[:0]
	for _, g := range groups {
		if err := g.prepare(); err != nil {
			a.Log.Errorf("Dropping %d metrics: %v", len(g.metrics), err)
			continue
		}
		valid = append(valid, g)
	}
	if len(valid) == 0 {
		// If no valid group remains, drop all metrics
		return nil
	}

	payload := bytes.NewBuffer(make([]byte, 0, a.lastPayloadSize))
	var writer io.Writer = payload

	// Wrap with gzip writer if compression is enabled
	if a.ContentEncoding == "gzip" {
		if a.gzipWriter == nil {
			a.gzipWriter = gzip.NewWriter(payload)
		} else {
			a.gzipWriter.Reset(payload)
		}
		writer = a.gzipWriter
	}

	// Stream the MessagePack data directly to the writer (gzipped or not).
	// A single measurement is sent as a map, multiple measurements as an array
	// of maps.
	msgpWriter := msgp.NewWriter(writer)
	if len(valid) > 1 {
		if err := msgpWriter.WriteArrayHeader(uint32(len(valid))); err != nil {
			return fmt.Errorf("marshalling message failed: %w", err)
		}
	}
	for _, g := range valid {
		if err := g.encode(msgpWriter, a.nils); err != nil {
			return fmt.Errorf("marshalling message failed: %w", err)
		}
	}
	if err := msgpWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush MessagePack writer: %w", err)
	}

	// Close gzip writer before reading the buffer and release the buffer
	if a.ContentEncoding == "gzip" {
		err := a.gzipWriter.Close()
		a.gzipWriter.Reset(io.Discard)
		if err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}
	a.lastPayloadSize = payload.Len()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", a.URL, payload)
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
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
