package opcua

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/selfstat"
)

type readClientWorkarounds struct {
	UseUnregisteredReads bool `toml:"use_unregistered_reads"`
}

type readClientConfig struct {
	ReconnectErrorThreshold *uint64               `toml:"reconnect_error_threshold"`
	ReadRetryTimeout        config.Duration       `toml:"read_retry_timeout"`
	ReadRetries             uint64                `toml:"read_retry_count"`
	ReadClientWorkarounds   readClientWorkarounds `toml:"request_workarounds"`
	input.InputClientConfig
}

// readClient Requests the current values from the required nodes when gather is called.
type readClient struct {
	*input.OpcUAInputClient

	ReconnectErrorThreshold uint64
	ReadRetryTimeout        time.Duration
	ReadRetries             uint64
	ReadSuccess             selfstat.Stat
	ReadError               selfstat.Stat
	Workarounds             readClientWorkarounds

	// Internal flags
	reqIDs          []*ua.ReadValueID
	maxNodesPerRead int
	ctx             context.Context
	forceReconnect  bool
}

func (rc *readClientConfig) createReadClient(log telegraf.Logger) (*readClient, error) {
	inputClient, err := rc.InputClientConfig.CreateInputClient(log)
	if err != nil {
		return nil, err
	}

	// The polling reader has no subscriptions to preserve across reconnects, so
	// gopcua's auto-reconnect only races with our own connect cycle and floods
	// the server with sessions during outages. Let Telegraf own reconnection.
	inputClient.DisableAutoReconnect = true

	tags := map[string]string{
		"endpoint": inputClient.Config.OpcUAClientConfig.Endpoint,
	}

	if rc.ReadRetryTimeout == 0 {
		rc.ReadRetryTimeout = config.Duration(100 * time.Millisecond)
	}

	// Set default for ReconnectErrorThreshold if not configured
	// Use the default value of reconnect after every error and
	// allow the user to override that setting including forcing
	// a reconnect after every cycle by setting zero.
	reconnectThreshold := uint64(1)
	if rc.ReconnectErrorThreshold != nil {
		reconnectThreshold = *rc.ReconnectErrorThreshold
	}

	return &readClient{
		OpcUAInputClient:        inputClient,
		ReconnectErrorThreshold: reconnectThreshold,
		ReadRetryTimeout:        time.Duration(rc.ReadRetryTimeout),
		ReadRetries:             rc.ReadRetries,
		ReadSuccess:             selfstat.Register("opcua", "read_success", tags),
		ReadError:               selfstat.Register("opcua", "read_error", tags),
		Workarounds:             rc.ReadClientWorkarounds,
	}, nil
}

func (o *readClient) connect() error {
	o.ctx = context.Background()
	o.forceReconnect = false

	if err := o.OpcUAClient.Connect(o.ctx); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	// Fetch namespace array for namespace URI support
	if err := o.OpcUAClient.UpdateNamespaceArray(o.ctx); err != nil {
		o.Log.Warnf("Failed to fetch namespace array: %v", err)
		// Continue anyway - this is only needed if using namespace URIs
	}

	// Query the server-imposed limit on nodes per read request so large node
	// sets can be split accordingly. The property is optional and zero means
	// "no limit"; in both cases all nodes are sent in a single request.
	o.maxNodesPerRead = 0
	limits, err := o.Client.Read(o.ctx, &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{{
			NodeID: ua.NewNumericNodeID(0, id.Server_ServerCapabilities_OperationLimits_MaxNodesPerRead),
		}},
	})
	switch {
	case err != nil:
		o.Log.Debugf("Querying the server's read limit failed: %v", err)
	case len(limits.Results) == 1 && limits.Results[0].Status == ua.StatusOK && limits.Results[0].Value != nil:
		if v, ok := limits.Results[0].Value.Value().(uint32); ok && v > 0 {
			// Clamp to avoid overflowing int on 32-bit platforms
			o.maxNodesPerRead = int(min(v, math.MaxInt32))
			o.Log.Debugf("Server limits read requests to %d nodes", v)
		}
	default:
		o.Log.Debug("Server does not report a read limit")
	}

	// Browse-based discovery runs on every connect so server-side schema
	// changes (added or removed nodes, renumbered namespaces) are picked up
	// on reconnect. DiscoverNodes replaces the previously discovered groups
	// and InitNodeMetricMapping rebuilds the mapping from scratch.
	if len(o.Config.Browse.Paths) > 0 {
		if err := o.OpcUAInputClient.DiscoverNodes(o.ctx); err != nil {
			return fmt.Errorf("browse discovery failed: %w", err)
		}
		if err := o.OpcUAInputClient.InitNodeMetricMapping(); err != nil {
			return fmt.Errorf("initializing node metric mapping failed: %w", err)
		}
	}

	// Make sure we setup the node-ids correctly after reconnect
	// as the server might be restarted and IDs changed
	if err := o.OpcUAInputClient.InitNodeIDs(); err != nil {
		return fmt.Errorf("initializing node IDs failed: %w", err)
	}

	o.reqIDs = make([]*ua.ReadValueID, 0, len(o.NodeIDs))
	if o.Workarounds.UseUnregisteredReads {
		for _, nid := range o.NodeIDs {
			o.reqIDs = append(o.reqIDs, &ua.ReadValueID{NodeID: nid})
		}
	} else {
		regResp, err := o.Client.RegisterNodes(o.ctx, &ua.RegisterNodesRequest{
			NodesToRegister: o.NodeIDs,
		})
		if err != nil {
			return fmt.Errorf("registering nodes failed: %w", err)
		}

		for _, v := range regResp.RegisteredNodeIDs {
			o.reqIDs = append(o.reqIDs, &ua.ReadValueID{NodeID: v})
		}
	}

	if err := o.read(); err != nil {
		return fmt.Errorf("get data failed: %w", err)
	}

	return nil
}

func (o *readClient) ensureConnected() error {
	// Force reconnection if we had a session error in the previous cycle
	if o.forceReconnect || o.State() == opcua.Disconnected || o.State() == opcua.Closed {
		// If we're forcing a reconnection, but we're not in Disconnected state,
		// explicitly disconnect first
		if o.State() != opcua.Disconnected && o.State() != opcua.Closed {
			if err := o.Disconnect(context.Background()); err != nil {
				o.Log.Debug("Error while disconnecting: ", err)
			}
		}
		return o.connect()
	}
	return nil
}

func (o *readClient) currentValues() ([]telegraf.Metric, error) {
	connectStart := time.Now()
	if err := o.ensureConnected(); err != nil {
		return nil, err
	}
	o.Log.Tracef("Connection check took %s", time.Since(connectStart))

	if state := o.State(); state != opcua.Connected {
		return nil, fmt.Errorf("not connected, in state %q", state)
	}

	readStart := time.Now()
	if err := o.read(); err != nil {
		o.Log.Tracef("Read failed after %s: %v", time.Since(readStart), err)
		// We do not return the disconnect error, as this would mask the
		// original problem, but we do log it
		if derr := o.Disconnect(context.Background()); derr != nil {
			o.Log.Debug("Error while disconnecting: ", derr)
		}

		return nil, err
	}
	o.Log.Tracef("OPC UA read of %d nodes took %s", len(o.NodeIDs), time.Since(readStart))

	metricStart := time.Now()
	metrics := make([]telegraf.Metric, 0, len(o.NodeMetricMapping))
	var skipped int
	// Parse the resulting data into metrics
	for i := range o.NodeIDs {
		if !o.StatusCodeOK(o.LastReceivedData[i].Quality) {
			o.Log.Tracef("Skipping node %s: bad quality %v", o.NodeIDs[i], o.LastReceivedData[i].Quality)
			skipped++
			continue
		}

		metrics = append(metrics, o.MetricForNode(i))
	}
	o.Log.Tracef("Metric construction took %s (%d metrics, %d skipped due to bad quality)",
		time.Since(metricStart), len(metrics), skipped)

	return metrics, nil
}

func (o *readClient) read() error {
	// Split the nodes into multiple requests if the server limits the number
	// of nodes per read; a single request holds everything otherwise.
	var batches [][]*ua.ReadValueID
	if o.maxNodesPerRead > 0 {
		for chunk := range slices.Chunk(o.reqIDs, o.maxNodesPerRead) {
			batches = append(batches, chunk)
		}
	} else {
		batches = [][]*ua.ReadValueID{o.reqIDs}
	}

	var count uint64

	for {
		count++

		// Try to update the values for all registered nodes
		o.Log.Tracef("Sending %d OPC UA read request(s) for %d %s nodes (attempt %d)...",
			len(batches), len(o.reqIDs), nodeTypeLabel(o.Workarounds.UseUnregisteredReads), count)
		requestStart := time.Now()
		var err error
		var updated int
		for _, batch := range batches {
			var resp *ua.ReadResponse
			resp, err = o.Client.Read(o.ctx, &ua.ReadRequest{
				MaxAge:             2000,
				TimestampsToReturn: ua.TimestampsToReturnBoth,
				NodesToRead:        batch,
			})
			if err != nil {
				break
			}
			if len(resp.Results) != len(batch) {
				err = fmt.Errorf("server returned %d results for %d requested nodes", len(resp.Results), len(batch))
				break
			}
			for i, d := range resp.Results {
				o.UpdateNodeValue(updated+i, d)
			}
			updated += len(batch)
		}
		plcDuration := time.Since(requestStart)
		if err == nil {
			// Success, exit with all node values updated
			o.ReadSuccess.Incr(1)
			o.forceReconnect = false
			o.Log.Tracef("OPC UA read completed in %s, updated %d node values", plcDuration, updated)
			return nil
		}

		o.ReadError.Incr(1)
		o.Log.Tracef("OPC UA read request failed after %s: %v", plcDuration, err)

		isSessionError := errors.Is(err, ua.StatusBadSessionIDInvalid) ||
			errors.Is(err, ua.StatusBadSessionNotActivated) ||
			errors.Is(err, ua.StatusBadSecureChannelIDInvalid)

		// Flag session error for next cycle if encountered
		if isSessionError {
			o.forceReconnect = true
		}

		switch {
		case count > o.ReadRetries:
			// We exceeded the number of retries and should exit
			return fmt.Errorf("reading %s nodes failed after %d attempts: %w",
				nodeTypeLabel(o.Workarounds.UseUnregisteredReads), count, err)
		case isSessionError:
			// Retry after the defined period as session and channels should be refreshed
			o.Log.Debugf("reading failed with %v, retry %d / %d...", err, count, o.ReadRetries)
			time.Sleep(o.ReadRetryTimeout)
		default:
			// Non-retryable error, there is nothing we can do
			return fmt.Errorf("reading %s nodes failed: %w",
				nodeTypeLabel(o.Workarounds.UseUnregisteredReads), err)
		}
	}
}

// Helper function to provide more accurate error messages
func nodeTypeLabel(useUnregistered bool) string {
	if useUnregistered {
		return "unregistered"
	}
	return "registered"
}
