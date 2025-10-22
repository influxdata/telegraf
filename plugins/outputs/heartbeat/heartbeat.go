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
	"net/http"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/logger"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

// jsonSchemaVersion is the version of the JSON schema to be used for any data sent
const jsonSchemaVersion = 1

type Heartbeat struct {
	URL        config.Secret             `toml:"url"`
	InstanceID string                    `toml:"instance_id"`
	Token      config.Secret             `toml:"token"`
	Interval   config.Duration           `toml:"interval"`
	Include    []string                  `toml:"include"`
	Logs       LogsConfig                `toml:"logs"`
	Headers    map[string]*config.Secret `toml:"headers"`
	Log        telegraf.Logger           `toml:"-"`
	common_http.HTTPClientConfig

	client        *http.Client
	logCallbackID string
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	// Output message parts
	message   message
	logEvents []*logEvent

	// Statistics
	stats statistics

	sync.Mutex
}

type LogsConfig struct {
	Limit    uint64 `toml:"limit"`
	LogLevel string `toml:"level"`

	level telegraf.LogLevel
}

type statistics struct {
	metrics     atomic.Uint64
	logErrors   atomic.Uint64
	logWarnings atomic.Uint64

	lastUpdate       time.Time
	lastUpdateFailed bool
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

	// Construct the fixed part of the message
	h.message = message{
		ID:      h.InstanceID,
		Version: internal.FormatFullVersion(),
		Schema:  jsonSchemaVersion,
	}

	for _, inc := range h.Include {
		switch inc {
		case "configs", "metrics", "logs":
			// Do nothing, those are valid
		case "hostname":
			host, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("getting hostname failed: %w", err)
			}
			h.message.Hostname = host
		case "log-details":
			// Set default log level if necessary
			if h.Logs.LogLevel == "" {
				h.Logs.LogLevel = "error"
			}
			h.Logs.level = telegraf.LogLevelFromString(h.Logs.LogLevel)
			if h.Logs.level == telegraf.None && h.Logs.LogLevel != "" && h.Logs.LogLevel != "none" {
				return fmt.Errorf("invalid log-level %q", h.Logs.LogLevel)
			}
		case "status":
			//			h.Log.Warn("'include' setting 'status' currently only return 'OK'")
		default:
			return fmt.Errorf("invalid 'include' setting %q", inc)
		}
	}

	return nil
}

func (h *Heartbeat) Connect() error {
	// Make sure we register a logging callback if we need to collect logs
	if (slices.Contains(h.Include, "logs") || slices.Contains(h.Include, "log-details")) && h.logCallbackID == "" {
		h.logCallbackID = logger.AddCallback(h.handleLogEvent)
	}
	h.stats.lastUpdate = time.Now()

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

		for {
			select {
			case <-cctx.Done():
				return
			case <-ticker.C:
				if err := h.send(); err != nil {
					h.Log.Error(err)
				}
			}
		}
	}(ctx)

	return nil
}

func (h *Heartbeat) Close() error {
	if h.logCallbackID != "" {
		logger.RemoveCallback(h.logCallbackID)
	}

	if h.cancel != nil {
		h.cancel()
	}

	if h.client != nil {
		h.client.CloseIdleConnections()
	}

	return nil
}

func (h *Heartbeat) Write(metrics []telegraf.Metric) error {
	h.stats.metrics.Add(uint64(len(metrics)))

	return nil
}

func (h *Heartbeat) send() error {
	// Snapshot the current information state for sending the message
	h.Lock()
	count := h.stats.metrics.Load()
	logErrs := h.stats.logErrors.Load()
	logWarns := h.stats.logWarnings.Load()
	logEvents := h.logEvents
	var lastUpdate int64
	if h.stats.lastUpdateFailed {
		lastUpdate = h.stats.lastUpdate.Unix()
	}
	h.Unlock()

	// Add the last successful update timestamp if any previous update failed
	if lastUpdate > 0 {
		h.message.LastSuccessfulUpdate = &lastUpdate
	} else {
		h.message.LastSuccessfulUpdate = nil
	}

	// Construct the message
	for _, item := range h.Include {
		switch item {
		case "metrics":
			h.message.Metrics = &count
		case "configs":
			h.message.ConfigSources = &config.Sources
		case "logs":
			if h.message.Logs == nil {
				h.message.Logs = &logsMessage{}
			}
			h.message.Logs.Errors = &logErrs
			h.message.Logs.Warnings = &logWarns
		case "log-details":
			if h.message.Logs == nil {
				h.message.Logs = &logsMessage{}
			}

			var entries []logEntry
			if h.Logs.Limit == 0 || h.Logs.Limit > uint64(len(h.logEvents)) {
				entries = getLogEntriesUnlimited(logEvents)
			} else {
				entries = getLogEntriesLimited(logEvents, int(h.Logs.Limit))
			}
			h.message.Logs.Entries = &entries
		case "status":
			// TODO: Evaluate the status condition
			h.message.Status = "OK"
		}
	}

	// Create the message body
	var body bytes.Buffer
	data, err := json.Marshal(h.message)
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.stats.lastUpdateFailed = true

		// Read the response body in case of any error
		response, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			return fmt.Errorf("received status %d (%s) with decoding message failed: %w", resp.StatusCode, resp.Status, rerr)
		}
		return fmt.Errorf("received status %d (%s) with message %s", resp.StatusCode, resp.Status, response)
	}

	// Update statistics on successful sent
	h.Lock()
	h.stats.lastUpdate = time.Now()
	h.stats.lastUpdateFailed = false
	h.stats.metrics.Add(-count)
	h.stats.logErrors.Add(-logErrs)
	h.stats.logWarnings.Add(-logWarns)
	h.logEvents = h.logEvents[len(logEvents):]
	h.Unlock()

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
