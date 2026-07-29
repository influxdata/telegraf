package cisco_telemetry_mdt

import (
	"encoding/json"
	"errors"
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

func (s *parser) parse(
	grouper *metric.SeriesGrouper,
	content *telemetry.TelemetryField,
	encodingPath string,
	isDeleted bool,
	tags map[string]string,
	timestamp time.Time,
) []error {
	// Do alias lookup, to shorten measurement names
	measurement := encodingPath
	if alias, ok := s.aliases[encodingPath]; ok {
		measurement = alias
	} else {
		s.Lock()
		if !s.warned[encodingPath] {
			s.log.Debugf("No measurement alias for encoding path: %s", encodingPath)
			s.warned[encodingPath] = true
		}
		s.Unlock()
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
		extraTags:   s.extraTags,
		propMap:     s.propMap,
		nxpathMap:   s.nxpathMap[encodingPath],
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
	if s.includeDelete {
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

func (s *state) parseField(field *telemetry.TelemetryField, prefix string, tags map[string]string, timestamp time.Time) []error {
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
		s.grouper.Add(s.measurement, tags, timestamp, name, value)

		return nil
	}

	// Get extra-tags defined by the user in embedded_tags
	extraTags := s.extraTags[s.tagPrefix+name]
	fields := make([]*telemetry.TelemetryField, 0, len(field.Fields))
	for _, subfield := range field.Fields {
		if _, isExtraTag := extraTags[subfield.Name]; isExtraTag {
			tags[name+"/"+strings.ReplaceAll(subfield.Name, "-", "_")] = decodeTag(subfield)
		} else {
			fields = append(fields, subfield)
		}
	}

	// No more fields to process
	if len(fields) == 0 {
		return nil
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

func (s *state) parseClassAttributeField(field *telemetry.TelemetryField, tags map[string]string, timestamp time.Time) []error {
	if s.path == "rib" {
		// handle native data path rib
		s.parseRib(field.Fields, tags, timestamp)
		return nil
	}
	if s.path == "microburst" {
		s.parseMicroburst(field.Fields, tags, timestamp)
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

	// Delete dn tag to prevent it from being added to the next node's attributes
	delete(tags, "dn")

	return errs
}

// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
func (s *state) parseDME(fields []*telemetry.TelemetryField, prefix string, tags map[string]string, timestamp time.Time) []error {
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

	if len(rn) > 0 {
		tags[prefix] = rn
	} else if !dn { // Check for distinguished name being present
		return []error{errors.New("failed while decoding NX-OS: missing 'dn' field")}
	}

	for _, subfield := range fields {
		if subfield.Name != "rn" {
			errs = append(errs, s.parseField(subfield, "", tags, timestamp)...)
		}
	}
	delete(tags, prefix)

	return errs
}

func (s *state) parseEvents(events []*telemetry.TelemetryField, prefix string, tags map[string]string, timestamp time.Time) []error {
	var attrs *telemetry.TelemetryField
	if events[0] != nil && len(events[0].Fields) >= 2 {
		if events[0].Fields[0].Name == "subscriptionId" {
			attrs = events[0].Fields[1].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
		} else if events[0].Fields[1].Name == "subscriptionId" {
			attrs = events[0].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
		}
	}

	// Parse the attribute subfields according to their class
	if attrs == nil {
		//nolint:prealloc // We expect errors to be empty by default
		var errs []error
		for _, sub := range events {
			errs = append(errs, s.parseClassAttributeField(sub, tags, timestamp)...)
		}
		return errs
	}

	return s.parseDME(attrs.Fields, prefix, tags, timestamp)
}

// NXAPI structure: https://developer.cisco.com/docs/cisco-nexus-9000-series-nx-api-cli-reference-release-9-2x/
func (s *state) parseNXRows(rows []*telemetry.TelemetryField, prefix string, tags map[string]string, timestamp time.Time) []error {
	var errs []error

	for _, row := range rows {
		if len(row.Fields) == 0 {
			continue
		}

		tags[prefix] = decodeTag(row.Fields[0])
		for i, field := range row.Fields {
			if i == 0 { // First subfield contains the index, promote it from value to tag
				// We can have subfield so recursively handle it.
				if len(row.Fields) == 1 {
					tags["row_number"] = strconv.FormatInt(int64(i), 10)
					errs = append(errs, s.parseField(field, "", tags, timestamp)...)
				}
			} else {
				errs = append(errs, s.parseField(field, "", tags, timestamp)...)
			}
			// Nxapi we can't identify keys always from prefix
			tags["row_number"] = strconv.FormatInt(int64(i), 10)
		}
		delete(tags, prefix)
	}

	return errs
}

func (s *state) parseRib(fields []*telemetry.TelemetryField, tags map[string]string, timestamp time.Time) {
	for _, subfield := range fields {
		// For Every table fill the keys which are vrfName, address and masklen
		switch subfield.Name {
		case "vrfName", "address", "maskLen":
			tags[subfield.Name] = decodeTag(subfield)
		}
		if value := decode(subfield); value != nil {
			s.grouper.Add(s.path, tags, timestamp, subfield.Name, value)
		}
		if subfield.Name != "nextHop" {
			continue
		}
		// For next hop table fill the keys in the tag - which is address and vrfname
		for _, subf := range subfield.Fields {
			for _, ff := range subf.Fields {
				switch ff.Name {
				case "address", "vrfName":
					key := "nextHop/" + ff.Name
					tags[key] = decodeTag(ff)
				}
				if value := decode(ff); value != nil {
					name := "nextHop/" + ff.Name
					s.grouper.Add(s.path, tags, timestamp, name, value)
				}
			}
		}
	}
}

func (s *state) parseMicroburst(fields []*telemetry.TelemetryField, tags map[string]string, timestamp time.Time) {
	if len(fields) < 3 || len(fields[2].Fields) == 0 || len(fields[2].Fields[0].Fields) < 4 {
		return
	}

	root := fields[2].Fields[0].Fields[3]
	for _, subfield := range root.Fields {
		if subfield.Name == "interfaceName" {
			tags[subfield.Name] = decodeTag(subfield)
		}

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
				s.grouper.Add(s.path, tags, timestamp, subf.Name, value)
			}
		}
	}
}
