package gnmi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/proto/gnmi_ext"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/gnmi/extensions/jnpr_gnmi_extention"
	"github.com/influxdata/telegraf/plugins/common/yangmodel"
)

// Define the warning to show if we cannot get a metric name.
const emptyNameWarning = `Got empty metric-name for response (field %q), usually
indicating configuration issues as the response cannot be related to any
subscription.Please open an issue on https://github.com/influxdata/telegraf
including your device model and the following response data:
%+v
This message is only printed once.`

type HandlerConfig struct {
	EmitDeleteMetrics             bool     `toml:"emit_delete_metrics"`
	CanonicalFieldNames           bool     `toml:"canonical_field_names"`
	TrimSlash                     bool     `toml:"trim_field_names"`
	TagPathPrefix                 bool     `toml:"prefix_tag_key_with_path"`
	GuessPathStrategy             string   `toml:"path_guessing_strategy"`
	VendorExt                     []string `toml:"vendor_specific"`
	YangModelPaths                []string `toml:"yang_model_paths"`
	DefaultName                   string   `toml:"-"`
	EnforceFirstNamespaceAsOrigin bool     `toml:"-"`
}

func (cfg *HandlerConfig) Check() error {
	// Check vendor_specific options configured by user
	if err := choice.CheckSlice(cfg.VendorExt, supportedExtensions); err != nil {
		return fmt.Errorf("unsupported vendor_specific option: %w", err)
	}

	// Check path guessing and handle deprecated option
	switch cfg.GuessPathStrategy {
	case "", "none", "common path", "subscription":
	default:
		return fmt.Errorf("invalid 'path_guessing_strategy' %q", cfg.GuessPathStrategy)
	}

	// Load the YANG models if specified by the user and check for errors early.
	// We redo this later to actually use the decode.
	if len(cfg.YangModelPaths) > 0 {
		if _, err := yangmodel.NewDecoder(cfg.YangModelPaths...); err != nil {
			return fmt.Errorf("invalid YANG model decoder: %w", err)
		}
	}

	return nil
}

func (cfg *HandlerConfig) Handler(log telegraf.Logger) (*Handler, error) {
	h := &Handler{
		HandlerConfig: cfg,
		aliases:       make(map[*pathInfo]string),
		tagStore:      newTagStore(),
		log:           log,
	}

	// Load the YANG models if specified by the user
	if len(cfg.YangModelPaths) > 0 {
		decoder, err := yangmodel.NewDecoder(cfg.YangModelPaths...)
		if err != nil {
			return nil, fmt.Errorf("creating YANG model decoder failed: %w", err)
		}
		h.decoder = decoder
	}
	return h, nil
}

type Handler struct {
	*HandlerConfig
	log telegraf.Logger

	// Internal
	aliases            map[*pathInfo]string
	decoder            *yangmodel.Decoder
	tagStore           *tagStore
	emptyNameWarnShown bool

	tagSubscriptions []*TagSubscription
}

func (h *Handler) AddTagSubscription(s *TagSubscription) {
	h.tagSubscriptions = append(h.tagSubscriptions, s)
	h.tagStore.add(s)
}

func (h *Handler) Process(acc telegraf.Accumulator, source string, response *gnmi.SubscribeResponse) {
	if h.log.Level().Includes(telegraf.Trace) {
		buf, err := protojson.Marshal(response)
		if err != nil {
			h.log.Debugf("Marshalling response failed: %v", err)
		} else {
			t := response.GetUpdate().GetTimestamp()
			h.log.Debugf("Got update_%v: %s", t, string(buf))
		}
	}

	r, ok := response.Response.(*gnmi.SubscribeResponse_Update)
	if !ok {
		return
	}

	update := r.Update

	// Extract the metadata
	timestamp, tags, prefix := h.handleUpdateMetadata(source, update, response.GetExtension())

	// Handle "update" notifications contained in the response
	h.handleUpdates(acc, update.Update, timestamp, tags, prefix)

	// Handle "delete" notifications contained in the response if requested
	if h.EmitDeleteMetrics {
		h.handleDeletes(acc, update.Delete, timestamp, tags, prefix)
	}
}

func (h *Handler) handleUpdateMetadata(
	source string,
	notification *gnmi.Notification,
	extension []*gnmi_ext.Extension,
) (time.Time, map[string]string, *pathInfo) {
	timestamp := time.Unix(0, notification.Timestamp)

	// Extract tags from potential extension in the update notification
	headerTags := map[string]string{"source": source}

	for _, ext := range extension {
		currentExt := ext.GetRegisteredExt().Msg
		if currentExt == nil {
			break
		}

		switch ext.GetRegisteredExt().Id {
		case eidJuniperTelemetryHeader:
			// Juniper Header extension
			// Decode it only if user requested it
			if choice.Contains("juniper_header", h.VendorExt) {
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
	if h.EnforceFirstNamespaceAsOrigin {
		prefix.enforceFirstNamespaceAsOrigin()
	}
	return timestamp, headerTags, prefix
}

// Handle SubscribeResponse_Update message from gNMI and parse contained telemetry data
func (h *Handler) handleUpdates(acc telegraf.Accumulator, updates []*gnmi.Update, timestamp time.Time, headerTags map[string]string, prefix *pathInfo) {
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
		if h.EnforceFirstNamespaceAsOrigin {
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
		for key, val := range fullPath.tags(h.TagPathPrefix) {
			tags[key] = val
		}

		// TODO: Handle each field individually to allow in-JSON tags
		var tagUpdate bool
		for _, tagSub := range h.tagSubscriptions {
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
	var path string
	if !prefix.empty() {
		path = prefix.fullPath()
	} else if h.GuessPathStrategy == "common path" {
		paths := make([]*pathInfo, 0, len(valueFields))
		for _, f := range valueFields {
			paths = append(paths, f.path)
		}
		if prefixPath := guessPrefix(paths); prefixPath != "" {
			path = prefixPath
		}
	}

	// Parse individual update message and create measurements
	for _, field := range valueFields {
		if field.path.empty() {
			continue
		}

		// Prepare tags from prefix
		fieldTags := field.path.tags(h.TagPathPrefix)
		tags := make(map[string]string, len(headerTags)+len(fieldTags))
		for key, val := range headerTags {
			tags[key] = val
		}
		if path != "" {
			tags["path"] = path
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
			if h.DefaultName != "" {
				name = h.DefaultName
			} else {
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
		}

		aliasInfo := newInfoFromString(aliasPath)
		if h.EnforceFirstNamespaceAsOrigin {
			aliasInfo.enforceFirstNamespaceAsOrigin()
		}

		if tags["path"] == "" && h.GuessPathStrategy == "subscription" {
			tags["path"] = aliasInfo.String()
		}

		// Group metrics
		var key string
		if h.CanonicalFieldNames {
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
		if h.TrimSlash {
			key = strings.TrimLeft(key, "/.")
		}
		if key == "" {
			h.log.Errorf("Invalid empty path %q with alias %q", field.path.String(), aliasPath)
			continue
		}
		grouper.Add(name, tags, timestamp, key, field.value)
	}

	// Add grouped measurements
	for _, m := range grouper.Metrics() {
		acc.AddMetric(m)
	}
}

func (h *Handler) handleDeletes(acc telegraf.Accumulator, deletes []*gnmi.Path, timestamp time.Time, headerTags map[string]string, prefix *pathInfo) {
	paths := make([]*pathInfo, 0, len(deletes))
	for _, del := range deletes {
		if del == nil {
			continue
		}

		if len(del.Elem) == 0 && prefix.empty() {
			continue
		}

		fullPath := prefix.append(del)
		if h.EnforceFirstNamespaceAsOrigin {
			prefix.enforceFirstNamespaceAsOrigin()
		}
		if del.Origin != "" {
			fullPath.origin = del.Origin
		}
		paths = append(paths, fullPath)
	}

	// Some devices do not provide a prefix, so do some guesswork based
	// on the paths of the fields
	var path string
	if !prefix.empty() {
		path = prefix.fullPath()
	} else if h.GuessPathStrategy == "common path" {
		if prefixPath := guessPrefix(paths); prefixPath != "" {
			path = prefixPath
		}
	}

	// Parse individual update message and create measurements
	for _, field := range paths {
		if field.empty() {
			continue
		}

		// Prepare tags from prefix
		fieldTags := field.tags(h.TagPathPrefix)
		tags := make(map[string]string, len(headerTags)+len(fieldTags)+1)
		for key, val := range headerTags {
			tags[key] = val
		}
		if path != "" {
			tags["path"] = path
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
			if h.DefaultName != "" {
				name = h.DefaultName
			} else {
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
		}

		aliasInfo := newInfoFromString(aliasPath)
		if h.EnforceFirstNamespaceAsOrigin {
			aliasInfo.enforceFirstNamespaceAsOrigin()
		}

		if tags["path"] == "" && h.GuessPathStrategy == "subscription" {
			tags["path"] = aliasInfo.String()
		}

		fields := map[string]interface{}{"operation": "delete"}
		acc.AddFields(name, fields, tags, timestamp)
	}
}
