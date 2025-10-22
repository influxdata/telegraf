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

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/vmihailenco/msgpack/v5"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultURL         = "http://localhost:8000/api/v1/write/msgpack"
	defaultTimeout     = 5 * time.Second
	defaultContentType = "application/msgpack"
	defaultUserAgent   = "Telegraf-Arc-Output-Plugin"
	defaultBatchSize   = 1000
)

// Arc output plugin for writing metrics to Arc time-series database using MessagePack binary protocol
type Arc struct {
	// Arc MessagePack API URL
	URL string `toml:"url"`

	// HTTP timeout
	Timeout config.Duration `toml:"timeout"`

	// API Key for authentication
	APIKey config.Secret `toml:"api_key"`

	// Database name for multi-database architecture
	// Routes metrics to a specific database namespace (e.g., "production", "staging", "default")
	Database string `toml:"database"`

	// HTTP Headers
	Headers map[string]string `toml:"headers"`

	// Content encoding: "gzip" or "identity" (default: gzip)
	ContentEncoding string `toml:"content_encoding"`

	// User agent string
	UserAgent string `toml:"user_agent"`

	// Batch size for MessagePack writes
	BatchSize int `toml:"batch_size"`

	// Log
	Log telegraf.Logger `toml:"-"`

	client *http.Client
}

// ArcColumnarData represents columnar format data for Arc's MessagePack format
type ArcColumnarData struct {
	Measurement string                 `msgpack:"m"`
	Columns     map[string]interface{} `msgpack:"columns"`
}

func (*Arc) SampleConfig() string {
	return sampleConfig
}

func (a *Arc) Init() error {
	// Set defaults
	if a.URL == "" {
		a.URL = defaultURL
	}

	if a.Timeout == 0 {
		a.Timeout = config.Duration(defaultTimeout)
	}

	if a.UserAgent == "" {
		a.UserAgent = defaultUserAgent
	}

	if a.ContentEncoding == "" {
		a.ContentEncoding = "gzip"
	}

	if a.BatchSize == 0 {
		a.BatchSize = defaultBatchSize
	}

	return nil
}

func (a *Arc) Connect() error {
	a.client = &http.Client{
		Timeout: time.Duration(a.Timeout),
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	// Try to construct health endpoint
	healthURL := "http://localhost:8000/health"
	if a.URL != "" {
		// Parse URL and construct health endpoint
		healthURL = a.URL[:len(a.URL)-len("/api/v1/write/msgpack")] + "/health"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		a.Log.Warnf("Unable to check Arc health endpoint: %v", err)
		return nil // Don't fail on health check
	}

	resp, err := a.client.Do(req)
	if err != nil {
		a.Log.Warnf("Arc health check failed: %v (continuing anyway)", err)
		return nil // Don't fail connection if health check fails
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		a.Log.Info("Successfully connected to Arc (MessagePack binary protocol)")
	}

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

	// Group metrics by measurement name for columnar format
	measurementGroups := make(map[string][]telegraf.Metric)
	for _, metric := range metrics {
		name := metric.Name()
		measurementGroups[name] = append(measurementGroups[name], metric)
	}

	// Convert each measurement group to columnar format
	columnarData := make([]ArcColumnarData, 0, len(measurementGroups))

	for measurementName, metricsGroup := range measurementGroups {
		if len(metricsGroup) == 0 {
			continue
		}

		// Initialize columns map
		columns := make(map[string]interface{})

		// Create slices for each column
		timestamps := make([]int64, len(metricsGroup))

		// Track all unique field and tag keys
		fieldKeys := make(map[string]bool)
		tagKeys := make(map[string]bool)

		// First pass: collect all unique keys
		for _, metric := range metricsGroup {
			for _, field := range metric.FieldList() {
				fieldKeys[field.Key] = true
			}
			for _, tag := range metric.TagList() {
				tagKeys[tag.Key] = true
			}
		}

		// Initialize field and tag columns
		fieldColumns := make(map[string][]interface{})
		for key := range fieldKeys {
			fieldColumns[key] = make([]interface{}, len(metricsGroup))
		}

		tagColumns := make(map[string][]string)
		for key := range tagKeys {
			tagColumns[key] = make([]string, len(metricsGroup))
		}

		// Second pass: populate columns
		for i, metric := range metricsGroup {
			// Add timestamp
			timestamps[i] = metric.Time().UnixMilli()

			// Add fields
			fieldMap := make(map[string]interface{})
			for _, field := range metric.FieldList() {
				fieldMap[field.Key] = field.Value
			}
			for key := range fieldKeys {
				if val, ok := fieldMap[key]; ok {
					fieldColumns[key][i] = val
				} else {
					fieldColumns[key][i] = nil
				}
			}

			// Add tags
			tagMap := make(map[string]string)
			for _, tag := range metric.TagList() {
				tagMap[tag.Key] = tag.Value
			}
			for key := range tagKeys {
				if val, ok := tagMap[key]; ok {
					tagColumns[key][i] = val
				} else {
					tagColumns[key][i] = ""
				}
			}
		}

		// Build columns map
		columns["time"] = timestamps

		// Add all field columns
		for key, values := range fieldColumns {
			columns[key] = values
		}

		// Add all tag columns
		for key, values := range tagColumns {
			columns[key] = values
		}

		columnarData = append(columnarData, ArcColumnarData{
			Measurement: measurementName,
			Columns:     columns,
		})
	}

	// Serialize with MessagePack
	// If there's only one measurement, send it directly; otherwise send as array
	var payload []byte
	var err error

	if len(columnarData) == 1 {
		payload, err = msgpack.Marshal(columnarData[0])
	} else {
		payload, err = msgpack.Marshal(columnarData)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal MessagePack: %w", err)
	}

	// Compress if enabled
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

	// Prepare request
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", a.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", defaultContentType)
	req.Header.Set("User-Agent", a.UserAgent)

	// Add API key if provided
	if !a.APIKey.Empty() {
		apiKey, err := a.APIKey.Get()
		if err == nil {
			req.Header.Set("x-api-key", apiKey.String())
			apiKey.Destroy()
		}
	}

	// Add database header for multi-database routing
	if a.Database != "" {
		req.Header.Set("x-arc-database", a.Database)
	}

	// Content encoding
	if a.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// Add custom headers
	for k, v := range a.Headers {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to write to Arc: %w", err)
	}
	defer resp.Body.Close()

	// Check response (Arc returns 204 No Content on success)
	if resp.StatusCode != 204 && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("arc returned status %d (failed to read body: %w)", resp.StatusCode, err)
		}
		return fmt.Errorf("arc returned status %d: %s", resp.StatusCode, string(body))
	}

	a.Log.Debugf("Successfully wrote %d metrics to Arc via MessagePack", len(metrics))
	return nil
}

func init() {
	outputs.Add("arc", func() telegraf.Output {
		return &Arc{
			Timeout:         config.Duration(defaultTimeout),
			ContentEncoding: "gzip",
			BatchSize:       defaultBatchSize,
		}
	})
}
