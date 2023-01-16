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
)

type handler struct {
	address            string
	aliases            map[string]string
	tagsubs            []TagSubscription
	maxMsgSize         int
	emptyNameWarnShown bool
	tagStore           *tagNode
	log                telegraf.Logger
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
	prefixTags["path"] = prefix

	// Process and remove tag-only updates from the response
	for i := len(response.Update.Update) - 1; i >= 0; i-- {
		update := response.Update.Update[i]
		fullPath := pathWithPrefix(response.Update.Prefix, update.Path)
		for _, tagSub := range h.tagsubs {
			if equalPathNoKeys(fullPath, tagSub.fullPath) {
				h.log.Debugf("Tag-subscription update for %q: %+v", tagSub.Name, update)
				h.storeTags(update, tagSub)
				response.Update.Update = append(response.Update.Update[:i], response.Update.Update[i+1:]...)
			}
		}
	}

	// Parse individual Update message and create measurements
	var name, lastAliasPath string
	for _, update := range response.Update.Update {
		fullPath := pathWithPrefix(response.Update.Prefix, update.Path)

		// Prepare tags from prefix
		tags := make(map[string]string, len(prefixTags))
		for key, val := range prefixTags {
			tags[key] = val
		}
		aliasPath, fields := h.handleTelemetryField(update, tags, prefix)

		if tagOnlyTags := h.checkTags(fullPath); tagOnlyTags != nil {
			for k, v := range tagOnlyTags {
				if alias, ok := h.aliases[k]; ok {
					tags[alias] = fmt.Sprint(v)
				} else {
					tags[k] = fmt.Sprint(v)
				}
			}
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

type tagNode struct {
	elem     *gnmiLib.PathElem
	tagName  string
	value    *gnmiLib.TypedValue
	tagStore map[string][]*tagNode
}

type tagResults struct {
	names  []string
	values []*gnmiLib.TypedValue
}

func (w *handler) storeTags(update *gnmiLib.Update, sub TagSubscription) {
	updateKeys := pathKeys(update.Path)
	var foundKey bool
	for _, requiredKey := range sub.Elements {
		foundKey = false
		for _, elem := range updateKeys {
			if elem.Name == requiredKey {
				foundKey = true
			}
		}
		if !foundKey {
			return
		}
	}
	// All required keys present for this TagSubscription
	w.tagStore.insert(updateKeys, sub.Name, update.Val)
}

func (node *tagNode) insert(keys []*gnmiLib.PathElem, name string, value *gnmiLib.TypedValue) {
	if len(keys) == 0 {
		node.value = value
		node.tagName = name
		return
	}
	var found *tagNode
	key := keys[0]
	keyName := key.Name
	if node.tagStore == nil {
		node.tagStore = make(map[string][]*tagNode)
	}
	if _, ok := node.tagStore[keyName]; !ok {
		node.tagStore[keyName] = make([]*tagNode, 0)
	}
	for _, node := range node.tagStore[keyName] {
		if compareKeys(node.elem.Key, key.Key) {
			found = node
			break
		}
	}
	if found == nil {
		found = &tagNode{elem: keys[0]}
		node.tagStore[keyName] = append(node.tagStore[keyName], found)
	}
	found.insert(keys[1:], name, value)
}

func (node *tagNode) retrieve(keys []*gnmiLib.PathElem, tagResults *tagResults) {
	if node.value != nil {
		tagResults.names = append(tagResults.names, node.tagName)
		tagResults.values = append(tagResults.values, node.value)
	}
	for _, key := range keys {
		if elems, ok := node.tagStore[key.Name]; ok {
			for _, node := range elems {
				if compareKeys(node.elem.Key, key.Key) {
					node.retrieve(keys, tagResults)
				}
			}
		}
	}
}

func (w *handler) checkTags(fullPath *gnmiLib.Path) map[string]interface{} {
	results := &tagResults{}
	w.tagStore.retrieve(pathKeys(fullPath), results)
	tags := make(map[string]interface{})
	for idx := range results.names {
		vals, _ := gnmiToFields(results.names[idx], results.values[idx])
		for k, v := range vals {
			tags[k] = v
		}
	}
	return tags
}
