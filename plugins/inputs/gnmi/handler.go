package gnmi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/proto/gnmi_ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/yangmodel"
	"github.com/influxdata/telegraf/plugins/inputs/gnmi/extensions/jnpr_gnmi_extention"
	"github.com/influxdata/telegraf/selfstat"
)

const eidJuniperTelemetryHeader = 1

type handler struct {
	host                          string
	port                          string
	aliases                       map[*pathInfo]string
	tagsubs                       []tagSubscription
	maxMsgSize                    int
	emptyNameWarnShown            bool
	vendorExt                     []string
	emitDeleteMetrics             bool
	tagStore                      *tagStore
	trace                         bool
	canonicalFieldNames           bool
	trimSlash                     bool
	tagPathPrefix                 bool
	guessPathStrategy             string
	decoder                       *yangmodel.Decoder
	enforceFirstNamespaceAsOrigin bool
	log                           telegraf.Logger
	keepalive.ClientParameters
}

// SubscribeGNMI and extract telemetry data
func (h *handler) subscribeGNMI(ctx context.Context, acc telegraf.Accumulator, tlscfg *tls.Config, request *gnmi.SubscribeRequest) error {
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

	if h.ClientParameters.Time > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(h.ClientParameters))
	}

	// Used to report the status of the TCP connection to the device. If the
	// GNMI connection goes down, but TCP is still up this will still report
	// connected until the TCP connection times out.
	connectStat := selfstat.Register("gnmi", "grpc_connection_status", map[string]string{"source": h.host})
	defer connectStat.Set(0)

	address := net.JoinHostPort(h.host, h.port)
	client, err := grpc.NewClient(address, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	subscribeClient, err := gnmi.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %w", err)
	}

	// If io.EOF is returned, the stream may have ended and stream status
	// can be determined by calling Recv.
	if err := subscribeClient.Send(request); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to send subscription request: %w", err)
	}
	connectStat.Set(1)
	h.log.Debugf("Connection to gNMI device %s established", address)

	defer h.log.Debugf("Connection to gNMI device %s closed", address)
	for ctx.Err() == nil {
		var reply *gnmi.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if !errors.Is(err, io.EOF) && ctx.Err() == nil {
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

		response, ok := reply.Response.(*gnmi.SubscribeResponse_Update)
		if !ok {
			continue
		}

		// Extract the metadata
		timestamp, tags, prefix := h.handleUpdateMetadata(response.Update, reply.GetExtension())

		// Handle "update" notifications contained in the response
		h.handleUpdates(acc, response.Update.Update, timestamp, tags, prefix)

		// Handle "delete" notifications contained in the response if requested
		if h.emitDeleteMetrics {
			h.handleDeletes(acc, response.Update.Delete, timestamp, tags, prefix)
		}
	}
	return nil
}

func (h *handler) handleUpdateMetadata(notification *gnmi.Notification, extension []*gnmi_ext.Extension) (time.Time, map[string]string, *pathInfo) {
	timestamp := time.Unix(0, notification.Timestamp)

	// Extract tags from potential extension in the update notification
	headerTags := make(map[string]string)
	for _, ext := range extension {
		currentExt := ext.GetRegisteredExt().Msg
		if currentExt == nil {
			break
		}

		switch ext.GetRegisteredExt().Id {
		case eidJuniperTelemetryHeader:
			// Juniper Header extension
			// Decode it only if user requested it
			if choice.Contains("juniper_header", h.vendorExt) {
				juniperHeader := &jnpr_gnmi_extention.GnmiJuniperTelemetryHeaderExtension{}
				if err := proto.Unmarshal(currentExt, juniperHeader); err != nil {
					h.log.Errorf("unmarshal gnmi Juniper Header extension failed: %v", err)
				} else {
					// Add only relevant Tags from the Juniper Header extension.
					// These are required for aggregation
					headerTags["component_id"] = strconv.FormatUint(uint64(juniperHeader.GetComponentId()), 10)
					headerTags["component"] = juniperHeader.GetComponent()
					headerTags["sub_component_id"] = strconv.FormatUint(uint64(juniperHeader.GetSubComponentId()), 10)
				}
			}
		default:
			continue
		}
	}

	// Extract the path part valid for the whole set of updates if any
	prefix := newInfoFromPath(notification.Prefix)
	if h.enforceFirstNamespaceAsOrigin {
		prefix.enforceFirstNamespaceAsOrigin()
	}

	// Add info to the tags
	headerTags["source"] = h.host
	if !prefix.empty() {
		headerTags["path"] = prefix.fullPath()
	}

	return timestamp, headerTags, prefix
}

// Handle SubscribeResponse_Update message from gNMI and parse contained telemetry data
func (h *handler) handleUpdates(acc telegraf.Accumulator, updates []*gnmi.Update, timestamp time.Time, headerTags map[string]string, prefix *pathInfo) {
	grouper := metric.NewSeriesGrouper()

	// Process and remove tag-updates from the response first so we can
	// add all available tags to the metrics later.
	var valueFields []updateField
	for _, update := range updates {
		if update.Path == nil {
			continue
		}

		if len(update.Path.Elem) == 0 && prefix.empty() {
			continue
		}

		fullPath := prefix.append(update.Path)
		if h.enforceFirstNamespaceAsOrigin {
			prefix.enforceFirstNamespaceAsOrigin()
		}
		if update.Path.Origin != "" {
			fullPath.origin = update.Path.Origin
		}

		fields, err := h.newFieldsFromUpdate(fullPath, update)
		if err != nil {
			h.log.Errorf("Processing update %v failed: %v", update, err)
		}

		// Prepare tags from prefix
		tags := make(map[string]string, len(headerTags))
		for key, val := range headerTags {
			tags[key] = val
		}
		for key, val := range fullPath.tags(h.tagPathPrefix) {
			tags[key] = val
		}

		// TODO: Handle each field individually to allow in-JSON tags
		var tagUpdate bool
		for _, tagSub := range h.tagsubs {
			if !fullPath.equalsPathNoKeys(tagSub.fullPath) {
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
			valueFields = append(valueFields, fields...)
		}
	}

	// Some devices do not provide a prefix, so do some guesswork based
	// on the paths of the fields
	if headerTags["path"] == "" && h.guessPathStrategy == "common path" {
		paths := make([]*pathInfo, 0, len(valueFields))
		for _, f := range valueFields {
			paths = append(paths, f.path)
		}
		if prefixPath := guessPrefix(paths); prefixPath != "" {
			headerTags["path"] = prefixPath
		}
	}

	// Parse individual update message and create measurements
	for _, field := range valueFields {
		if field.path.empty() {
			continue
		}

		// Prepare tags from prefix
		fieldTags := field.path.tags(h.tagPathPrefix)
		tags := make(map[string]string, len(headerTags)+len(fieldTags))
		for key, val := range headerTags {
			tags[key] = val
		}
		for key, val := range fieldTags {
			tags[key] = val
		}

		// Add the tags derived via tag-subscriptions
		for k, v := range h.tagStore.lookup(field.path, tags) {
			tags[k] = v
		}

		// Lookup alias for the metric
		aliasPath, name := h.lookupAlias(field.path)
		if name == "" {
			h.log.Debugf("No measurement alias for gNMI path: %s", field.path)
			if !h.emptyNameWarnShown {
				if buf, err := json.Marshal(updates); err == nil {
					h.log.Warnf(emptyNameWarning, field.path, string(buf))
				} else {
					h.log.Warnf(emptyNameWarning, field.path, updates)
				}
				h.emptyNameWarnShown = true
			}
		}

		aliasInfo := newInfoFromString(aliasPath)
		if h.enforceFirstNamespaceAsOrigin {
			aliasInfo.enforceFirstNamespaceAsOrigin()
		}

		if tags["path"] == "" && h.guessPathStrategy == "subscription" {
			tags["path"] = aliasInfo.String()
		}

		// Group metrics
		var key string
		if h.canonicalFieldNames {
			// Strip the origin is any for the field names
			field.path.origin = ""
			key = field.path.String()
			key = strings.ReplaceAll(key, "-", "_")
		} else {
			// If the alias is a subpath of the field path and the alias is
			// shorter than the full path to avoid an empty key, then strip the
			// common part of the field is prefixed with the alias path. Note
			// the origins can match or be empty and be considered equal.
			if relative := aliasInfo.relative(field.path, true); relative != "" {
				key = relative
			} else {
				// Otherwise use the last path element as the field key
				key = field.path.base()
			}
			key = strings.ReplaceAll(key, "-", "_")
		}
		if h.trimSlash {
			key = strings.TrimLeft(key, "/.")
		}
		if key == "" {
			h.log.Errorf("Invalid empty path %q with alias %q", field.path.String(), aliasPath)
			continue
		}
		grouper.Add(name, tags, timestamp, key, field.value)
	}

	// Add grouped measurements
	for _, metricToAdd := range grouper.Metrics() {
		acc.AddMetric(metricToAdd)
	}
}

func (h *handler) handleDeletes(acc telegraf.Accumulator, deletes []*gnmi.Path, timestamp time.Time, headerTags map[string]string, prefix *pathInfo) {
	paths := make([]*pathInfo, 0, len(deletes))
	for _, del := range deletes {
		if del == nil {
			continue
		}

		if len(del.Elem) == 0 && prefix.empty() {
			continue
		}

		fullPath := prefix.append(del)
		if h.enforceFirstNamespaceAsOrigin {
			prefix.enforceFirstNamespaceAsOrigin()
		}
		if del.Origin != "" {
			fullPath.origin = del.Origin
		}
		paths = append(paths, fullPath)
	}

	// Some devices do not provide a prefix, so do some guesswork based
	// on the paths of the fields
	if headerTags["path"] == "" && h.guessPathStrategy == "common path" {
		if prefixPath := guessPrefix(paths); prefixPath != "" {
			headerTags["path"] = prefixPath
		}
	}

	// Parse individual update message and create measurements
	for _, field := range paths {
		if field.empty() {
			continue
		}

		// Prepare tags from prefix
		fieldTags := field.tags(h.tagPathPrefix)
		tags := make(map[string]string, len(headerTags)+len(fieldTags))
		for key, val := range headerTags {
			tags[key] = val
		}
		for key, val := range fieldTags {
			tags[key] = val
		}

		// Add the tags derived via tag-subscriptions
		for k, v := range h.tagStore.lookup(field, tags) {
			tags[k] = v
		}

		// Lookup alias for the metric
		aliasPath, name := h.lookupAlias(field)
		if name == "" {
			h.log.Debugf("No measurement alias for gNMI path: %s", field)
			if !h.emptyNameWarnShown {
				if buf, err := json.Marshal(deletes); err == nil {
					h.log.Warnf(emptyNameWarning, field, string(buf))
				} else {
					h.log.Warnf(emptyNameWarning, field, deletes)
				}
				h.emptyNameWarnShown = true
			}
		}

		aliasInfo := newInfoFromString(aliasPath)
		if h.enforceFirstNamespaceAsOrigin {
			aliasInfo.enforceFirstNamespaceAsOrigin()
		}

		if tags["path"] == "" && h.guessPathStrategy == "subscription" {
			tags["path"] = aliasInfo.String()
		}

		fields := map[string]interface{}{"operation": "delete"}

		fmt.Printf("[delete] got name %q with tags %+v\n", name, tags)

		acc.AddFields(name, fields, tags, timestamp)
	}
}

// Try to find the alias for the given path
type aliasCandidate struct {
	path, alias string
}

func (h *handler) lookupAlias(info *pathInfo) (aliasPath, alias string) {
	candidates := make([]aliasCandidate, 0, len(h.aliases))
	for i, a := range h.aliases {
		if !i.isSubPathOf(info) {
			continue
		}
		candidates = append(candidates, aliasCandidate{i.String(), a})
	}
	if len(candidates) == 0 {
		return "", ""
	}

	// Reverse sort the candidates by path length so we can use the longest match
	sort.SliceStable(candidates, func(i, j int) bool {
		return len(candidates[i].path) > len(candidates[j].path)
	})

	return candidates[0].path, candidates[0].alias
}

func guessPrefix(paths []*pathInfo) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return paths[0].dir()
	}
	segments := make([]segment, 0, len(paths[0].segments))
	commonPath := &pathInfo{
		origin:   paths[0].origin,
		segments: append(segments, paths[0].segments...),
	}
	for _, f := range paths[1:] {
		commonPath.keepCommonPart(f)
	}
	if commonPath.empty() {
		return ""
	}
	return commonPath.String()
}
