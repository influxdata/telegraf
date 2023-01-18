package gnmi

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"path"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type handler struct {
	address            string
	aliases            map[string]string
	tagsubs            []TagSubscription
	maxMsgSize         int
	emptyNameWarnShown bool
	tagStore           *tagStore
	log                telegraf.Logger
}

func newHandler(addr string, aliases map[string]string, subs []TagSubscription, maxsize int, l telegraf.Logger) *handler {
	return &handler{
		address:    addr,
		aliases:    aliases,
		tagsubs:    subs,
		maxMsgSize: maxsize,
		tagStore:   newTagStore(subs),
		log:        l,
	}
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
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	subscribeClient, err := gnmiLib.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %w", err)
	}

	// If io.EOF is returned, the stream may have ended and stream status
	// can be determined by calling Recv.
	if err := subscribeClient.Send(request); err != nil && err != io.EOF {
		return fmt.Errorf("failed to send subscription request: %w", err)
	}

	h.log.Debugf("Connection to gNMI device %s established", h.address)
	defer h.log.Debugf("Connection to gNMI device %s closed", h.address)
	for ctx.Err() == nil {
		var reply *gnmiLib.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if err != io.EOF && ctx.Err() == nil {
				return fmt.Errorf("aborted gNMI subscription: %w", err)
			}
			break
		}

		buf, err := protojson.Marshal(reply)
		if err != nil {
			h.log.Debugf("marshal failed: %v", err)
		} else {
			t := reply.GetUpdate().GetTimestamp()
			h.log.Debugf("update_%v: %s", t, string(buf))
		}

		if response, ok := reply.Response.(*gnmiLib.SubscribeResponse_Update); ok {
			h.handleSubscribeResponseUpdate(acc, response)
		}
	}
	return nil
}

// Handle SubscribeResponse_Update message from gNMI and parse contained telemetry data
func (h *handler) handleSubscribeResponseUpdate(acc telegraf.Accumulator, response *gnmiLib.SubscribeResponse_Update) {
	var prefix, prefixAliasPath string
	grouper := metric.NewSeriesGrouper()
	timestamp := time.Unix(0, response.Update.Timestamp)
	prefixTags := make(map[string]string)

	if response.Update.Prefix != nil {
		var err error
		if prefix, prefixAliasPath, err = handlePath(response.Update.Prefix, prefixTags, h.aliases, ""); err != nil {
			h.log.Errorf("handling path %q failed: %v", response.Update.Prefix, err)
		}
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
			if err := h.tagStore.insert(tagSub, fullPath, fields); err != nil {
				h.log.Errorf("inserting tag failed: %w", err)
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

		// Add the tags derived via tag-subscriptions
		for k, v := range h.tagStore.lookup(fullPath) {
			tags[k] = v
		}

		aliasPath, fields := h.handleTelemetryField(update, tags, prefix)

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
		}

		// Check for empty names
		if name == "" && !h.emptyNameWarnShown {
			h.log.Warnf(emptyNameWarning, response.Update)
			h.emptyNameWarnShown = true
		}

		// Group metrics
		for k, v := range fields {
			key := k
			if len(aliasPath) < len(key) && len(aliasPath) != 0 {
				// This may not be an exact prefix, due to naming style
				// conversion on the key.
				key = key[len(aliasPath)+1:]
			} else if len(aliasPath) >= len(key) {
				// Otherwise use the last path element as the field key.
				key = path.Base(key)

				// If there are no elements skip the item; this would be an
				// invalid message.
				key = strings.TrimLeft(key, "/.")
				if key == "" {
					h.log.Errorf("invalid empty path: %q", k)
					continue
				}
			}
			grouper.Add(name, tags, timestamp, key, v)
		}

		lastAliasPath = aliasPath
	}

	// Add grouped measurements
	for _, metricToAdd := range grouper.Metrics() {
		acc.AddMetric(metricToAdd)
	}
}

// HandleTelemetryField and add it to a measurement
func (h *handler) handleTelemetryField(update *gnmiLib.Update, tags map[string]string, prefix string) (string, map[string]interface{}) {
	gpath, aliasPath, err := handlePath(update.Path, tags, h.aliases, prefix)
	if err != nil {
		h.log.Errorf("handling path %q failed: %v", update.Path, err)
	}
	fields, err := gnmiToFields(strings.Replace(gpath, "-", "_", -1), update.Val)
	if err != nil {
		h.log.Errorf("error parsing update value %q: %v", update.Val, err)
	}
	return aliasPath, fields
}
