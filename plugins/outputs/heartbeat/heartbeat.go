//go:generate ../../../tools/readme_config_includer/generator
package heartbeat

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Heartbeat struct {
	URL        config.Secret             `toml:"url"`
	InstanceID string                    `toml:"instance_id"`
	Token      config.Secret             `toml:"token"`
	Interval   config.Duration           `toml:"interval"`
	Include    []string                  `toml:"include"`
	Headers    map[string]*config.Secret `toml:"headers"`
	Log        telegraf.Logger           `toml:"-"`
	common_http.HTTPClientConfig

	client *http.Client
	cancel context.CancelFunc
	wg     sync.WaitGroup

	message map[string]interface{}
	metrics atomic.Uint64
}

func (*Heartbeat) SampleConfig() string {
	return sampleConfig
}

func (h *Heartbeat) Init() error {
	// Check settings
	if h.URL.Empty() {
		return errors.New("url required")
	}

	if h.InstanceID == "" {
		return errors.New("instance ID required")
	}

	if h.Interval <= 0 {
		return errors.New("invalid interval")
	}

	for _, inc := range h.Include {
		switch inc {
		case "configs", "hostname", "metrics":
			// Do nothing, those are valid
		case "logs":
			return fmt.Errorf("'include' setting %q not implemented yet", inc)
		case "status":
			h.Log.Warn("'include' setting 'status' currently only return 'OK'")
		default:
			return fmt.Errorf("invalid 'include' setting %q", inc)
		}
	}

	// Construct the fixed part of the message
	h.message = map[string]interface{}{
		"id":      h.InstanceID,
		"version": internal.FormatFullVersion(),
	}
	if slices.Contains(h.Include, "hostname") {
		host, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("getting hostname failed: %w", err)
		}
		h.message["hostname"] = host
	}

	return nil
}

func (h *Heartbeat) Connect() error {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	// Create the HTTP client
	client, err := h.HTTPClientConfig.CreateClient(ctx, h.Log)
	if err != nil {
		return fmt.Errorf("creating HTTP client failed: %w", err)
	}
	h.client = client

	// Start the ticker for sending heartbeat messages
	h.wg.Add(1)
	go func(cctx context.Context) {
		defer h.wg.Done()

		// Create a ticker for sending the messages with the given interval
		ticker := time.NewTicker(time.Duration(h.Interval))
		defer ticker.Stop()

		select {
		case <-cctx.Done():
			return
		case <-ticker.C:
			if err := h.send(); err != nil {
				h.Log.Error(err)
			}
		}
	}(ctx)

	return nil
}

func (h *Heartbeat) Close() error {
	if h.cancel != nil {
		h.cancel()
	}

	if h.client != nil {
		h.client.CloseIdleConnections()
	}

	return nil
}

func (h *Heartbeat) Write(metrics []telegraf.Metric) error {
	h.metrics.Add(uint64(len(metrics)))

	return nil
}

func (h *Heartbeat) send() error {
	// Get the number of metrics for optional sending
	count := h.metrics.Swap(0)

	// Construct the message
	message := maps.Clone(h.message)
	if slices.Contains(h.Include, "metrics") {
		message["metrics"] = count
	}
	if slices.Contains(h.Include, "configs") {
		message["configurations"] = config.Sources
	}
	if slices.Contains(h.Include, "logs") {
		// TODO: Retrive this information from the agent
		return errors.New("not supported yet")
	}
	if slices.Contains(h.Include, "status") {
		// TODO: Evaluate the status condition
		message["status"] = "OK"
	}

	// Create the message body
	var body bytes.Buffer
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("encoding message failed: %w", err)
	}
	if _, err := body.Write(data); err != nil {
		return fmt.Errorf("buffering message failed: %w", err)
	}

	// Construct the request
	urlRaw, err := h.URL.Get()
	if err != nil {
		return fmt.Errorf("getting URL secret failed: %w", err)
	}
	url := urlRaw.String()
	urlRaw.Destroy()

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return err
	}

	// Construct the header
	req.Header = http.Header(make(map[string][]string, len(h.Headers)+3))
	req.Header.Set("User-Agent", internal.ProductToken())
	req.Header.Add("Content-Type", "application/json")
	for k, raw := range h.Headers {
		v, err := raw.Get()
		if err != nil {
			return fmt.Errorf("getting %q secret failed: %w", k, err)
		}
		req.Header.Add(k, v.String())
		v.Destroy()
	}

	// Set the authentication if any
	if !h.Token.Empty() {
		token, err := h.Token.Get()
		if err != nil {
			return fmt.Errorf("getting token secret failed: %w", err)
		}
		req.Header.Add("Authentication", "Bearer "+token.String())
		token.Destroy()
	}

	// Send the message
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending message failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body in case of any error
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("writing to %q failed: %w", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received status %d (%s) with message %s", resp.StatusCode, resp.Status, response)
	}

	return nil
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &Heartbeat{
			Include:  []string{"hostname"},
			Interval: config.Duration(time.Minute),
		}
	})
}
