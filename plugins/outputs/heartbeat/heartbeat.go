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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/cel-go/cel"

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
	Status     StatusConfig              `toml:"status"`
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
	statuses  []program

	// Statistics
	stats             statistics
	statusInitialized atomic.Bool

	sync.Mutex
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
		case "configs", "statistics":
			// Do nothing, those are valid
		case "hostname":
			host, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("getting hostname failed: %w", err)
			}
			h.message.Hostname = host
		case "logs":
			// Set default log level if necessary
			if h.Logs.LogLevel == "" {
				h.Logs.LogLevel = "error"
			}
			h.Logs.level = telegraf.LogLevelFromString(h.Logs.LogLevel)
			if h.Logs.level == telegraf.None && h.Logs.LogLevel != "" && h.Logs.LogLevel != "none" {
				return fmt.Errorf("invalid log-level %q", h.Logs.LogLevel)
			}
		case "status":
			// Check the default value
			switch h.Status.Default {
			case "":
				h.Status.Default = "ok"
			case "ok", "warn", "fail", "undefined":
				// Do nothing, those are valid
			default:
				return fmt.Errorf("invalid status 'default' value %q", h.Status.Default)
			}
			h.Status.Default = strings.ToUpper(h.Status.Default)

			// Check the initial value
			switch h.Status.Initial {
			case "":
				h.statusInitialized.Store(true)
			case "ok", "warn", "fail", "undefined":
				// Do nothing, those are valid
			default:
				return fmt.Errorf("invalid status 'initial' value %q", h.Status.Initial)
			}
			h.Status.Initial = strings.ToUpper(h.Status.Initial)

			// Make sure the order is valid
			if len(h.Status.Order) == 0 {
				h.Status.Order = []string{"ok", "warn", "fail"}
			}
			seen := make(map[string]bool, 3)
			for _, o := range h.Status.Order {
				if seen[o] {
					return fmt.Errorf("duplicate value %q in status 'order'", o)
				}
				seen[o] = true
			}
			if h.Status.Ok != "" && !seen["ok"] {
				h.Log.Warn("condition for status \"ok\" will be ignored as it is not in the 'order' list")
			}
			if h.Status.Warn != "" && !seen["warn"] {
				h.Log.Warn("condition for status \"warn\" will be ignored as it is not in the 'order' list")
			}
			if h.Status.Fail != "" && !seen["fail"] {
				h.Log.Warn("condition for status \"fail\" will be ignored as it is not in the 'order' list")
			}

			// Compile the status programs
			if err := h.compileStatuses(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid 'include' setting %q", inc)
		}
	}

	return nil
}

func (h *Heartbeat) Connect() error {
	// Make sure we register a logging callback if we need to collect logs
	if (slices.Contains(h.Include, "logs") || slices.Contains(h.Include, "statistics")) && h.logCallbackID == "" {
		id, err := logger.AddCallback(h.handleLogEvent)
		if err != nil {
			return fmt.Errorf("registering logging callback failed: %w", err)
		}
		h.logCallbackID = id
	}

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

	// Wait for the heartbeat goroutine to finish before closing connections
	h.wg.Wait()

	if h.client != nil {
		h.client.CloseIdleConnections()
	}

	return nil
}

func (h *Heartbeat) Write(metrics []telegraf.Metric) error {
	h.stats.Lock()
	defer h.stats.Unlock()

	h.stats.metrics += uint64(len(metrics))
	h.statusInitialized.Store(true)

	// Heartbeat plugin does not process metrics; it sends heartbeats independently
	return nil
}

func (h *Heartbeat) send() error {
	// Snapshot the current information state for sending the message
	snapshot := h.stats.snapshot()
	h.Lock()
	logEvents := h.logEvents
	h.Unlock()

	// Add the last successful update timestamp if any previous update failed
	if snapshot.lastUpdateFailed {
		ts := snapshot.lastUpdate.Unix()
		h.message.LastSuccessfulUpdate = &ts
	} else {
		h.message.LastSuccessfulUpdate = nil
	}

	// Construct the message
	for _, item := range h.Include {
		switch item {
		case "statistics":
			h.message.Statistics = &statsEntry{
				Errors:   snapshot.logErrors,
				Warnings: snapshot.logWarnings,
				Metrics:  snapshot.metrics,
			}
		case "configs":
			sources := config.GetSources()
			h.message.ConfigSources = &sources
		case "logs":
			var entries []logEntry
			if h.Logs.Limit == 0 || h.Logs.Limit > uint64(len(h.logEvents)) {
				entries = getLogEntriesUnlimited(logEvents)
			} else {
				entries = getLogEntriesLimited(logEvents, int(h.Logs.Limit))
			}
			h.message.Logs = &entries
		case "status":
			if !h.statusInitialized.Load() {
				h.message.Status = h.Status.Initial
				continue
			}
			vars := snapshot.variables()
			h.message.Status = h.Status.Default
			for _, p := range h.statuses {
				match, err := p.eval(vars)
				if err != nil {
					return fmt.Errorf("evaluating status %q failed: %w", p.status, err)
				}
				if match {
					h.message.Status = p.status
					break
				}
			}
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
		req.Header.Add("Authorization", "Bearer "+token.String())
		token.Destroy()
	}

	// Send the message
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending message failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read the response body in case of any error
		response, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			return fmt.Errorf("received status %d (%s) with decoding message failed: %w", resp.StatusCode, resp.Status, rerr)
		}

		// Mark the statistics as not sent
		h.stats.Lock()
		h.stats.lastUpdateFailed = true
		h.stats.Unlock()

		return fmt.Errorf("received status %d (%s) with message %s", resp.StatusCode, resp.Status, response)
	}

	// Update statistics on successful sent
	h.stats.remove(snapshot, time.Now())
	h.Lock()
	h.logEvents = h.logEvents[len(logEvents):]
	h.Unlock()

	return nil
}

func (h *Heartbeat) compileStatuses() error {
	env, err := environment()
	if err != nil {
		return fmt.Errorf("creating status program environment failed: %w", err)
	}

	// Compile the programs
	h.statuses = make([]program, 0, len(h.Status.Order))
	for _, s := range h.Status.Order {
		// Get the expression
		var expression string
		switch s {
		case "ok":
			expression = h.Status.Ok
		case "warn":
			expression = h.Status.Warn
		case "fail":
			expression = h.Status.Fail
		default:
			return fmt.Errorf("invalid status 'order' value %q", s)
		}

		// Skip all empty expressions assuming they do not match
		if expression == "" {
			continue
		}

		// Compile the expression
		ast, issues := env.Compile(expression)
		if issues.Err() != nil {
			return fmt.Errorf("compiling expression for status %q failed: %w", s, issues.Err())
		}

		// Check if we got a boolean expression needed for filtering
		if ast.OutputType() != cel.BoolType {
			return fmt.Errorf("expression for status %q needs to return a boolean", s)
		}

		// Get the final program
		p, err := env.Program(ast, cel.EvalOptions(cel.OptOptimize))
		if err != nil {
			return fmt.Errorf("creating program for status %q failed: %w", s, err)
		}
		h.statuses = append(h.statuses, program{status: strings.ToUpper(s), prog: p})
	}
	return nil
}

func init() {
	outputs.Add("heartbeat", func() telegraf.Output {
		return &Heartbeat{
			Include:  []string{"hostname"},
			Interval: config.Duration(time.Minute),
		}
	})
}
