package gnmi

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"path"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/metric"
	jnprHeader "github.com/influxdata/telegraf/plugins/inputs/gnmi/extensions/jnpr_gnmi_extention"
	"github.com/influxdata/telegraf/selfstat"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	gnmiExt "github.com/openconfig/gnmi/proto/gnmi_ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const eidJuniperTelemetryHeader = 1

type handler struct {
	address             string
	aliases             map[string]string
	tagsubs             []TagSubscription
	maxMsgSize          int
	emptyNameWarnShown  bool
	vendorExt           []string
	tagStore            *tagStore
	trace               bool
	canonicalFieldNames bool
	trimSlash           bool
	log                 telegraf.Logger
}

// SubscribeGNMI and extract telemetry data
func (h *handler) subscribeGNMI(ctx context.Context, acc telegraf.Accumulator, tlscfg *tls.Config, request *gnmiLib.SubscribeRequest) error {
	var creds credentials.TransportCredentials
	if tlscfg != nil {
		creds = credentials.NewTLS(tlscfg)
	} else {
		creds = insecure.NewCredentials()
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	if h.maxMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(h.maxMsgSize),
		))
	}

	client, err := grpc.DialContext(ctx, h.address, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	subscribeClient, err := gnmiLib.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %w", err)
	}

	// If io.EOF is returned, the stream may have ended and stream status
	// can be determined by calling Recv.
	if err := subscribeClient.Send(request); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to send subscription request: %w", err)
	}

	h.log.Debugf("Connection to gNMI device %s established", h.address)

	// Used to report the status of the TCP connection to the device. If the
	// GNMI connection goes down, but TCP is still up this will still report
	// connected until the TCP connection times out.
	connectStat := selfstat.Register("gnmi", "grpc_connection_status", map[string]string{"source": h.address})
	connectStat.Set(1)

	defer h.log.Debugf("Connection to gNMI device %s closed", h.address)
	for ctx.Err() == nil {
		var reply *gnmiLib.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if !errors.Is(err, io.EOF) && ctx.Err() == nil {
				connectStat.Set(0)
				return fmt.Errorf("aborted gNMI subscription: %w", err)
			}
			break
		}

		if h.trace {
			buf, err := protojson.Marshal(reply)
			if err != nil {
				h.log.Debugf("Marshal failed: %v", err)
			} else {
				t := reply.GetUpdate().GetTimestamp()
				h.log.Debugf("Got update_%v: %s", t, string(buf))
			}
		}
		if response, ok := reply.Response.(*gnmiLib.SubscribeResponse_Update); ok {
			h.handleSubscribeResponseUpdate(acc, response, reply.GetExtension())
		}
	}

	connectStat.Set(0)
	return nil
}

// Handle SubscribeResponse_Update message from gNMI and parse contained telemetry data
func (h *handler) handleSubscribeResponseUpdate(acc telegraf.Accumulator, response *gnmiLib.SubscribeResponse_Update, extension []*gnmiExt.Extension) {
	var prefix, prefixAliasPath string
	grouper := metric.NewSeriesGrouper()
	timestamp := time.Unix(0, response.Update.Timestamp)
	prefixTags := make(map[string]string)

	// iter on each extension
	for _, ext := range extension {
		currentExt := ext.GetRegisteredExt().Msg
		if currentExt == nil {
			break
		}
		// extension ID
		switch ext.GetRegisteredExt().Id {
		// Juniper Header extention
		//EID_JUNIPER_TELEMETRY_HEADER = 1;
		case eidJuniperTelemetryHeader:
			// Decode it only if user requested it
			if choice.Contains("juniper_header", h.vendorExt) {
				juniperHeader := &jnprHeader.GnmiJuniperTelemetryHeaderExtension{}
				// unmarshal extention
				err := proto.Unmarshal(currentExt, juniperHeader)
				if err != nil {
					h.log.Errorf("unmarshal gnmi Juniper Header extension failed: %v", err)
					break
				}
				// Add only relevant Tags from the Juniper Header extention.
				// These are requiered for aggregation
				prefixTags["component_id"] = fmt.Sprint(juniperHeader.GetComponentId())
				prefixTags["component"] = fmt.Sprint(juniperHeader.GetComponent())
				prefixTags["sub_component_id"] = fmt.Sprint(juniperHeader.GetSubComponentId())
			}

		default:
			continue
		}
	}

	if response.Update.Prefix != nil {
		var origin string
		var err error
		if origin, prefix, prefixAliasPath, err = handlePath(response.Update.Prefix, prefixTags, h.aliases, ""); err != nil {
			h.log.Errorf("Handling path %q failed: %v", response.Update.Prefix, err)
		}
		prefix = origin + prefix
	}

	prefixTags["source"], _, _ = net.SplitHostPort(h.address)
	if prefix != "" {
		prefixTags["path"] = prefix
	}

	// Process and remove tag-updates from the response first so we will
	// add all available tags to the metrics later.
	var valueUpdates []*gnmiLib.Update
	for _, update := range response.Update.Update {
		fullPath := pathWithPrefix(response.Update.Prefix, update.Path)

		// Prepare tags from prefix
		tags := make(map[string]string, len(prefixTags))
		for key, val := range prefixTags {
			tags[key] = val
		}

		_, fields := h.handleTelemetryField(update, tags, prefix)

		var tagUpdate bool
		for _, tagSub := range h.tagsubs {
			if !equalPathNoKeys(fullPath, tagSub.fullPath) {
				continue
			}
			h.log.Debugf("Tag-subscription update for %q: %+v", tagSub.Name, update)
			if err := h.tagStore.insert(tagSub, fullPath, fields, tags); err != nil {
				h.log.Errorf("Inserting tag failed: %v", err)
			}
			tagUpdate = true
			break
		}
		if !tagUpdate {
			valueUpdates = append(valueUpdates, update)
		}
	}

	// Parse individual Update message and create measurements
	var name, lastAliasPath string
	for _, update := range valueUpdates {
		fullPath := pathWithPrefix(response.Update.Prefix, update.Path)

		// Prepare tags from prefix
		tags := make(map[string]string, len(prefixTags))
		for key, val := range prefixTags {
			tags[key] = val
		}

		aliasPath, fields := h.handleTelemetryField(update, tags, prefix)

		// Add the tags derived via tag-subscriptions
		for k, v := range h.tagStore.lookup(fullPath, tags) {
			tags[k] = v
		}

		// Inherent valid alias from prefix parsing
		if len(prefixAliasPath) > 0 && len(aliasPath) == 0 {
			aliasPath = prefixAliasPath
		}

		// Lookup alias if alias-path has changed
		if aliasPath != lastAliasPath {
			name = prefix
			if alias, ok := h.aliases[aliasPath]; ok {
				name = alias
			} else {
				h.log.Debugf("No measurement alias for gNMI path: %s", name)
			}
			lastAliasPath = aliasPath
		}

		// Check for empty names
		if name == "" && !h.emptyNameWarnShown {
			h.log.Warnf(emptyNameWarning, response.Update)
			h.emptyNameWarnShown = true
		}

		// Group metrics
		for k, v := range fields {
			key := k
			if h.canonicalFieldNames {
				// Strip the origin is any for the field names
				if parts := strings.SplitN(key, ":", 2); len(parts) == 2 {
					key = parts[1]
				}
			} else {
				if len(aliasPath) < len(key) && len(aliasPath) != 0 {
					// This may not be an exact prefix, due to naming style
					// conversion on the key.
					key = key[len(aliasPath)+1:]
				} else if len(aliasPath) >= len(key) {
					// Otherwise use the last path element as the field key.
					key = path.Base(key)
				}
			}
			if h.trimSlash {
				key = strings.TrimLeft(key, "/.")
			}
			if key == "" {
				h.log.Errorf("Invalid empty path: %q", k)
				continue
			}
			grouper.Add(name, tags, timestamp, key, v)
		}
	}

	// Add grouped measurements
	for _, metricToAdd := range grouper.Metrics() {
		acc.AddMetric(metricToAdd)
	}
}

// HandleTelemetryField and add it to a measurement
func (h *handler) handleTelemetryField(update *gnmiLib.Update, tags map[string]string, prefix string) (string, map[string]interface{}) {
	_, gpath, aliasPath, err := handlePath(update.Path, tags, h.aliases, prefix)
	if err != nil {
		h.log.Errorf("Handling path %q failed: %v", update.Path, err)
	}
	fields, err := gnmiToFields(strings.Replace(gpath, "-", "_", -1), update.Val)
	if err != nil {
		h.log.Errorf("Error parsing update value %q: %v", update.Val, err)
	}
	return aliasPath, fields
}
