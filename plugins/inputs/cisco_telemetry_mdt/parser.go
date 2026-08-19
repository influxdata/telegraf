package cisco_telemetry_mdt

import (
	"encoding/json"
	"errors"
	"maps"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type parser struct {
	includeDelete bool
	aliases       map[string]string
	extraTags     map[string]map[string]bool
	propMap       map[string]func(*telemetry.TelemetryField) interface{}
	nxpathMap     map[string]map[string]string // per path map

	log telegraf.Logger

	warned map[string]bool
	sync.Mutex
}

type state struct {
	path        string
	measurement string
	tagPrefix   string
	isNXOS      bool
	isEvent     bool

	extraTags map[string]map[string]bool
	propMap   map[string]func(*telemetry.TelemetryField) interface{}
	nxpathMap map[string]string // per path map

	grouper *metric.SeriesGrouper
}

func newParser(includeDelete bool, aliases, dmes map[string]string, embeddedTags []string, log telegraf.Logger) *parser {
	p := &parser{
		includeDelete: includeDelete,
		aliases:       make(map[string]string, len(aliases)),
		extraTags:     make(map[string]map[string]bool),
		propMap:       make(map[string]func(field *telemetry.TelemetryField) interface{}, len(dmes)+4),
		nxpathMap:     createDatabase(),
		warned:        make(map[string]bool),
		log:           log,
	}

	// Fill the parser aliases with the inverted list
	for alias, encodingPath := range aliases {
		p.aliases[encodingPath] = alias
	}

	// Fill extra tags
	for _, tag := range embeddedTags {
		dir := strings.ReplaceAll(path.Dir(tag), "-", "_")
		if _, found := p.extraTags[dir]; !found {
			p.extraTags[dir] = make(map[string]bool)
		}
		p.extraTags[dir][path.Base(tag)] = true
	}

	// Initialize property conversion map
	p.propMap["test"] = nxosValueXformUint64Toint64
	p.propMap["asn"] = nxosValueXformUint64ToString            // uint64 to string.
	p.propMap["subscriptionId"] = nxosValueXformUint64ToString // uint64 to string.
	p.propMap["operState"] = nxosValueXformUint64ToString      // uint64 to string.

	for dme, dmeKey := range dmes {
		switch dmeKey {
		case "uint64 to int":
			p.propMap[dme] = nxosValueXformUint64Toint64
		case "uint64 to string":
			p.propMap[dme] = nxosValueXformUint64ToString
		case "string to float64":
			p.propMap[dme] = nxosValueXformStringTofloat
		case "string to uint64":
			p.propMap[dme] = nxosValueXformStringToUint64
		case "string to int64":
			p.propMap[dme] = nxosValueXformStringToInt64
		case "auto-float-xfrom":
			p.propMap[dme] = nxosValueAutoXformFloatProp
		default:
			if !strings.HasPrefix(dme, "dnpath") {
				// Ignore non-path based property map
				continue
			}

			var payload nxPayloadXfromStructure
			if err := json.Unmarshal([]byte(dmeKey), &payload); err != nil {
				continue
			}

			// Build 2 level Hash nxpathMap Key = jsStruct.Name, Value = map of jsStruct.Prop
			// It will override the default of code if same path is provided in configuration.
			p.nxpathMap[payload.Name] = make(map[string]string, len(payload.Prop))
			for _, prop := range payload.Prop {
				p.nxpathMap[payload.Name][prop.Key] = prop.Value
			}
		}
	}

	return p
}

func (p *parser) parse(
	grouper *metric.SeriesGrouper,
	content *telemetry.TelemetryField,
	encodingPath string,
	isDeleted bool,
	tags map[string]string,
	timestamp time.Time,
) []error {
	// Do alias lookup, to shorten measurement names
	measurement := encodingPath
	if alias, ok := p.aliases[encodingPath]; ok {
		measurement = alias
	} else {
		p.Lock()
		if !p.warned[encodingPath] {
			p.log.Debugf("No measurement alias for encoding path: %s", encodingPath)
			p.warned[encodingPath] = true
		}
		p.Unlock()
	}

	// Determine what OS we are on and if the message encodes events
	// IOS-XR and IOS-XE have a colon in their encoding path, NX-OS does not
	isNXOS := !strings.ContainsRune(encodingPath, ':')
	isEvent := isNXOS && strings.Contains(encodingPath, "EVENT-LIST")

	// Initialize the parsing state
	state := &state{
		path:        encodingPath,
		measurement: measurement,
		tagPrefix:   strings.ReplaceAll(encodingPath, "-", "_") + "/",
		isNXOS:      isNXOS,
		isEvent:     isEvent,
		extraTags:   p.extraTags,
		propMap:     p.propMap,
		nxpathMap:   p.nxpathMap[encodingPath],
		grouper:     grouper,
	}

	// Parse the content if any
	var errs []error
	if content != nil {
		for _, subfield := range content.Fields {
			var prefix string
			switch subfield.Name {
			case "operation-metric":
				if len(subfield.Fields[0].Fields) > 0 {
					prefix = subfield.Fields[0].Fields[0].GetStringValue()
				}
			case "class-stats":
				if len(subfield.Fields[0].Fields) > 1 {
					prefix = subfield.Fields[0].Fields[1].GetStringValue()
				}
			}
			// Parse the content with and without prefix
			errs = append(errs, state.parseField(subfield, prefix, tags, timestamp)...)
			errs = append(errs, state.parseField(subfield, "", tags, timestamp)...)
		}
	}

	// Add a delete field if configured
	if p.includeDelete {
		grouper.Add(measurement, tags, timestamp, "delete", isDeleted)
	}

	return errs
}

// Recursively parse the "keys" element and convert to tags
func parseKeys(field *telemetry.TelemetryField, prefix string, tags map[string]string) {
	name := strings.ReplaceAll(field.Name, "-", "_")
	fullName := prefix
	if fullName != "" {
		fullName += "/"
	}
	fullName += name

	// Store the tag with the short-key if possible, otherwise use the full
	// tag-key containing the element path
	if value := decodeTag(field); name != "" && value != "" {
		if _, exists := tags[name]; !exists {
			tags[name] = value
		} else {
			tags[fullName] = value
		}
	}

	// Iterate over potential sub-elements
	for _, subfield := range field.Fields {
		parseKeys(subfield, fullName, tags)
	}
}

func (s *state) parseField(field *telemetry.TelemetryField, prefix string, parentTags map[string]string, timestamp time.Time) []error {
	name := strings.ReplaceAll(field.Name, "-", "_")

	// Exit early on fields to ignore
	value := decode(field)
	if (name == "modTs" || name == "createTs") && value == "never" {
		return nil
	}

	// Prefix the name if necessary
	if len(name) == 0 {
		name = prefix
	} else if prefix != "" {
		name = prefix + "/" + name
	}

	// Decode scalar fields if present and exit
	if value != nil {
		if s.isNXOS {
			// NXOS specific values take precedence if existing
			if val := nxosValueXform(field, s.propMap, s.nxpathMap); val != nil {
				value = val
			}
		}
		s.grouper.Add(s.measurement, parentTags, timestamp, name, value)

		return nil
	}

	// Get extra-tags defined by the user in embedded_tags
	tags := make(map[string]string)
	fields := make([]*telemetry.TelemetryField, 0, len(field.Fields))
	for _, subfield := range field.Fields {
		if _, isExtraTag := s.extraTags[s.tagPrefix+name][subfield.Name]; isExtraTag {
			tags[name+"/"+strings.ReplaceAll(subfield.Name, "-", "_")] = decodeTag(subfield)
		} else {
			fields = append(fields, subfield)
		}
	}

	// No more fields to process
	if len(fields) == 0 {
		return nil
	}

	// Prevent cloning on every node for performance reasons
	if len(tags) == 0 {
		tags = parentTags
	} else {
		// Copy over the parent tags, the extra tags take precedence
		for k, v := range parentTags {
			if _, found := tags[k]; !found {
				tags[k] = v
			}
		}
	}

	// Extract special field elements
	switch {
	case s.isNXOS && field.Name == "attributes":
		return s.parseDME(fields[0].Fields, prefix, tags, timestamp)
	case s.isNXOS && field.Name == "children" && !s.isEvent:
		// Parse the subfields according to their class
		//nolint:prealloc // We expect errors to be empty by default
		var errs []error
		for _, sub := range fields {
			errs = append(errs, s.parseClassAttributeField(sub, tags, timestamp)...)
		}
		return errs
	case s.isNXOS && field.Name == "children" && s.isEvent:
		return s.parseEvents(field.Fields, prefix, tags, timestamp)
	case s.isNXOS && strings.HasPrefix(field.Name, "ROW_"):
		return s.parseNXRows(fields, prefix, tags, timestamp)
	}

	// No special fields, continue with regular telemetry decoding of the tree
	//nolint:prealloc // We expect errors to be empty by default
	var errs []error
	for _, sub := range fields {
		errs = append(errs, s.parseField(sub, name, tags, timestamp)...)
	}
	return errs
}

func (s *state) parseClassAttributeField(field *telemetry.TelemetryField, parentTags map[string]string, timestamp time.Time) []error {
	switch s.path {
	case "rib":
		// handle native data path rib
		s.parseRib(field.Fields, parentTags, timestamp)
		return nil
	case "microburst":
		s.parseMicroburst(field.Fields, parentTags, timestamp)
		return nil
	}

	// Check if we have a DME structure
	// https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	if !strings.Contains(s.path, "sys/") {
		return nil
	}

	if field == nil || len(field.Fields) == 0 || len(field.Fields[0].Fields) == 0 || len(field.Fields[0].Fields[0].Fields) == 0 {
		return nil
	}

	if field.Fields[0].Fields[0].Fields[0].Name != "attributes" {
		return nil
	}

	nxAttributes := field.Fields[0].Fields[0].Fields[0].Fields[0]

	// Find dn tag among list of attributes
	tags := maps.Clone(parentTags)
	for _, subfield := range nxAttributes.Fields {
		if subfield.Name == "dn" {
			tags["dn"] = decodeTag(subfield)
			break
		}
	}

	// Add attributes to grouper with consistent dn tag
	var errs []error //nolint:prealloc // We expect the errors to be empty in most cases
	for _, subfield := range nxAttributes.Fields {
		errs = append(errs, s.parseField(subfield, "", tags, timestamp)...)
	}

	return errs
}

// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
func (s *state) parseDME(fields []*telemetry.TelemetryField, prefix string, parentTags map[string]string, timestamp time.Time) []error {
	var errs []error

	var rn string
	var dn bool
	for _, subfield := range fields {
		switch subfield.Name {
		case "rn":
			rn = decodeTag(subfield)
		case "dn":
			dn = true
		}
	}

	// Check for distinguished name being present
	tags := maps.Clone(parentTags)
	if rn != "" {
		tags[prefix] = rn
	} else if !dn {
		return []error{errors.New("failed while decoding NX-OS: missing 'dn' field")}
	}

	for _, subfield := range fields {
		if subfield.Name != "rn" {
			errs = append(errs, s.parseField(subfield, "", tags, timestamp)...)
		}
	}

	return errs
}

func (s *state) parseEvents(events []*telemetry.TelemetryField, prefix string, parentTags map[string]string, timestamp time.Time) []error {
	var attrs *telemetry.TelemetryField
	if events[0] != nil && len(events[0].Fields) >= 2 {
		var attrFields []*telemetry.TelemetryField
		if events[0].Fields[0].Name == "subscriptionId" && len(events[0].Fields[1].Fields) > 0 {
			attrFields = events[0].Fields[1].Fields
		} else if events[0].Fields[1].Name == "subscriptionId" {
			attrFields = events[0].Fields[0].Fields
		}
		valid := len(attrFields) > 0 && len(attrFields[0].Fields) > 0 && len(attrFields[0].Fields[0].Fields) > 0
		valid = valid && len(attrFields[0].Fields[0].Fields[0].Fields) > 0
		valid = valid && len(attrFields[0].Fields[0].Fields[0].Fields[0].Fields) > 0
		if valid {
			attrs = attrFields[0].Fields[0].Fields[0].Fields[0].Fields[0]
		}
	}

	// Parse the attribute subfields according to their class
	if attrs != nil {
		return s.parseDME(attrs.Fields, prefix, parentTags, timestamp)
	}

	//nolint:prealloc // We expect errors to be empty by default
	var errs []error
	for _, sub := range events {
		errs = append(errs, s.parseClassAttributeField(sub, parentTags, timestamp)...)
	}

	return errs
}

// NXAPI structure: https://developer.cisco.com/docs/cisco-nexus-9000-series-nx-api-cli-reference-release-9-2x/
func (s *state) parseNXRows(rows []*telemetry.TelemetryField, prefix string, parentTags map[string]string, timestamp time.Time) []error {
	var errs []error

	for i, row := range rows {
		if len(row.Fields) == 0 {
			continue
		}

		// First subfield contains the index, promote it from value to tag
		tags := maps.Clone(parentTags)
		tags[prefix] = decodeTag(row.Fields[0])
		tags["row_number"] = strconv.FormatInt(int64(i), 10)

		// If we only have a single column in the row, add it as a field too
		// to make sure the metric is emitted
		if len(row.Fields) == 1 {
			errs = append(errs, s.parseField(row.Fields[0], "", tags, timestamp)...)
			continue
		}

		// Add all other columns to the metric
		for _, field := range row.Fields[1:] {
			errs = append(errs, s.parseField(field, "", tags, timestamp)...)
		}
	}

	return errs
}

func (s *state) parseRib(fields []*telemetry.TelemetryField, parentTags map[string]string, timestamp time.Time) {
	// Collect the tags first as we need the complete set for assigning the
	// values to the correct series through the series grouper
	var nextHopFields []*telemetry.TelemetryField
	tags := maps.Clone(parentTags)
	metricFields := make(map[string]interface{}, len(fields))
	for _, subfield := range fields {
		switch subfield.Name {
		case "vrfName", "address", "maskLen":
			tags[subfield.Name] = decodeTag(subfield)
		case "nextHop":
			nextHopFields = subfield.Fields
		}

		if value := decode(subfield); value != nil {
			metricFields[subfield.Name] = value
		}
	}

	// Add the fields to the metrics using the full tag set
	for key, value := range metricFields {
		s.grouper.Add(s.measurement, tags, timestamp, key, value)
	}

	if len(nextHopFields) == 0 {
		return
	}

	// Now collect the next-hop information if any
	for _, subfield := range nextHopFields {
		hopTags := maps.Clone(tags)
		clear(metricFields)
		for _, subSubField := range subfield.Fields {
			switch subSubField.Name {
			case "address", "vrfName":
				hopTags["nextHop/"+subSubField.Name] = decodeTag(subSubField)
			default:
				if value := decode(subSubField); value != nil {
					metricFields["nextHop/"+subSubField.Name] = value
				}
			}
		}
		// Add the fields to the metrics using the full tag set
		for key, value := range metricFields {
			s.grouper.Add(s.measurement, hopTags, timestamp, key, value)
		}
	}
}

func (s *state) parseMicroburst(fields []*telemetry.TelemetryField, parentTags map[string]string, timestamp time.Time) {
	if len(fields) < 3 || len(fields[2].Fields) == 0 || len(fields[2].Fields[0].Fields) < 4 {
		return
	}

	root := fields[2].Fields[0].Fields[3]
	for _, subfield := range root.Fields {
		tags := maps.Clone(parentTags)
		if subfield.Name == "interfaceName" {
			tags[subfield.Name] = decodeTag(subfield)
		}

		// Collect the tags and metricFields first as the tags must be complete for
		// assigning the  field values to the correct series
		metricFields := make(map[string]interface{}, len(subfield.Fields))
		for _, subf := range subfield.Fields {
			switch subf.Name {
			case "sourceName":
				newstr := strings.Split(decodeTag(subf), "-[")
				if len(newstr) <= 2 {
					tags[subf.Name] = decodeTag(subf)
				} else {
					intfName := strings.Split(newstr[1], "]")
					queue := strings.Split(newstr[2], "]")
					tags["interface_name"] = intfName[0]
					tags["queue_number"] = queue[0]
				}
			case "startTs":
				tags[subf.Name] = decodeTag(subf)
			}
			if value := decode(subf); value != nil {
				metricFields[subf.Name] = value
			}
		}

		// Add the fields to the metrics using the full tag set
		for key, value := range metricFields {
			s.grouper.Add(s.measurement, tags, timestamp, key, value)
		}
	}
}
