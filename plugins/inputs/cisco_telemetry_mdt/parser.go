package cisco_telemetry_mdt

import (
	"errors"
	"strconv"
	"strings"
	"time"

	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"

	"github.com/influxdata/telegraf/metric"
)

// Recursively parse tag fields
func (c *CiscoTelemetryMDT) parseKeyField(tags map[string]string, field *telemetry.TelemetryField, prefix string) {
	localname := strings.ReplaceAll(field.Name, "-", "_")
	name := localname
	if len(localname) == 0 {
		name = prefix
	} else if len(prefix) > 0 {
		name = prefix + "/" + localname
	}

	if tag := decodeTag(field); len(name) > 0 && len(tag) > 0 {
		if _, exists := tags[localname]; !exists { // Use short keys whenever possible
			tags[localname] = tag
		} else {
			tags[name] = tag
		}
	}

	for _, subfield := range field.Fields {
		c.parseKeyField(tags, subfield, name)
	}
}

func (c *CiscoTelemetryMDT) parseContentField(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	prefix, encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	name := strings.ReplaceAll(field.Name, "-", "_")

	if (name == "modTs" || name == "createTs") && decode(field) == "never" {
		return
	}
	if len(name) == 0 {
		name = prefix
	} else if prefix != "" {
		name = prefix + "/" + name
	}

	extraTags := c.extraTags[strings.ReplaceAll(encodingPath, "-", "_")+"/"+name]
	if value := decode(field); value != nil {
		measurement := c.getMeasurementName(encodingPath)
		if val := c.nxosValueXform(field, encodingPath); val != nil {
			grouper.Add(measurement, tags, timestamp, name, val)
		} else {
			grouper.Add(measurement, tags, timestamp, name, value)
		}
		return
	}

	if len(extraTags) > 0 {
		for _, subfield := range field.Fields {
			if _, isExtraTag := extraTags[subfield.Name]; isExtraTag {
				tags[name+"/"+strings.ReplaceAll(subfield.Name, "-", "_")] = decodeTag(subfield)
			}
		}
	}

	var nxAttributes, nxChildren, nxRows *telemetry.TelemetryField
	isNXOS := !strings.ContainsRune(encodingPath, ':') // IOS-XR and IOS-XE have a colon in their encoding path, NX-OS does not
	isEVENT := isNXOS && strings.Contains(encodingPath, "EVENT-LIST")
	nxChildren = nil
	nxAttributes = nil
	for _, subfield := range field.Fields {
		if isNXOS && subfield.Name == "attributes" && len(subfield.Fields) > 0 {
			nxAttributes = subfield.Fields[0]
		} else if isNXOS && subfield.Name == "children" && len(subfield.Fields) > 0 {
			if !isEVENT {
				nxChildren = subfield
			} else {
				sub := subfield.Fields
				if len(sub) > 0 && sub[0] != nil && len(sub[0].Fields) >= 2 {
					if sub[0].Fields[0].Name == "subscriptionId" {
						nxAttributes = sub[0].Fields[1].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
					} else if sub[0].Fields[1].Name == "subscriptionId" {
						nxAttributes = sub[0].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
					}
				}
			}
			// if nxAttributes == NULL then class based query.
			if nxAttributes == nil {
				// call function walking over walking list.
				for _, sub := range subfield.Fields {
					c.parseClassAttributeField(grouper, sub, encodingPath, tags, timestamp)
				}
			}
		} else if isNXOS && strings.HasPrefix(subfield.Name, "ROW_") {
			nxRows = subfield
		} else if _, isExtraTag := extraTags[subfield.Name]; !isExtraTag { // Regular telemetry decoding
			c.parseContentField(grouper, subfield, name, encodingPath, tags, timestamp)
		}
	}

	if nxAttributes == nil && nxRows == nil {
		return
	}
	if nxRows != nil {
		// NXAPI structure: https://developer.cisco.com/docs/cisco-nexus-9000-series-nx-api-cli-reference-release-9-2x/
		for _, row := range nxRows.Fields {
			for i, subfield := range row.Fields {
				if i == 0 { // First subfield contains the index, promote it from value to tag
					tags[prefix] = decodeTag(subfield)
					// We can have subfield so recursively handle it.
					if len(row.Fields) == 1 {
						tags["row_number"] = strconv.FormatInt(int64(i), 10)
						c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
					}
				} else {
					c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
				}
				// Nxapi we can't identify keys always from prefix
				tags["row_number"] = strconv.FormatInt(int64(i), 10)
			}
			delete(tags, prefix)
		}
		return
	}

	// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	var rn string
	var dn bool
	for _, subfield := range nxAttributes.Fields {
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
		c.acc.AddError(errors.New("failed while decoding NX-OS: missing 'dn' field"))
		return
	}

	for _, subfield := range nxAttributes.Fields {
		if subfield.Name != "rn" {
			c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
		}
	}

	if nxChildren != nil {
		// This is a nested structure, children will inherit relative name keys of parent
		for _, subfield := range nxChildren.Fields {
			c.parseContentField(grouper, subfield, prefix, encodingPath, tags, timestamp)
		}
	}
	delete(tags, prefix)
}

func (c *CiscoTelemetryMDT) parseClassAttributeField(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	var nxAttributes *telemetry.TelemetryField
	isDme := strings.Contains(encodingPath, "sys/")
	if encodingPath == "rib" {
		// handle native data path rib
		parseRib(grouper, field, encodingPath, tags, timestamp)
		return
	}
	if encodingPath == "microburst" {
		// dump microburst
		parseMicroburst(grouper, field, encodingPath, tags, timestamp)
		return
	}
	if field == nil || !isDme || len(field.Fields) == 0 || len(field.Fields[0].Fields) == 0 || len(field.Fields[0].Fields[0].Fields) == 0 {
		return
	}

	if field.Fields[0] != nil && field.Fields[0].Fields != nil && field.Fields[0].Fields[0] != nil && field.Fields[0].Fields[0].Fields[0].Name != "attributes" {
		return
	}
	nxAttributes = field.Fields[0].Fields[0].Fields[0].Fields[0]

	// Find dn tag among list of attributes
	for _, subfield := range nxAttributes.Fields {
		if subfield.Name == "dn" {
			tags["dn"] = decodeTag(subfield)
			break
		}
	}
	// Add attributes to grouper with consistent dn tag
	for _, subfield := range nxAttributes.Fields {
		c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
	}
	// Delete dn tag to prevent it from being added to the next node's attributes
	delete(tags, "dn")
}

func (c *CiscoTelemetryMDT) getMeasurementName(encodingPath string) string {
	// Do alias lookup, to shorten measurement names
	measurement := encodingPath
	if alias, ok := c.internalAliases[encodingPath]; ok {
		measurement = alias
	} else {
		c.mutex.Lock()
		if !c.warned[encodingPath] {
			c.Log.Debugf("No measurement alias for encoding path: %s", encodingPath)
			c.warned[encodingPath] = true
		}
		c.mutex.Unlock()
	}
	return measurement
}

func parseMicroburst(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	var nxMicro *telemetry.TelemetryField
	var nxMicro1 *telemetry.TelemetryField
	// Microburst
	measurement := encodingPath
	if len(field.Fields) > 3 {
		nxMicro = field.Fields[2]
		if len(nxMicro.Fields) > 0 {
			nxMicro1 = nxMicro.Fields[0]
			if len(nxMicro1.Fields) >= 3 {
				nxMicro = nxMicro1.Fields[3]
			}
		}
	}
	for _, subfield := range nxMicro.Fields {
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
				grouper.Add(measurement, tags, timestamp, subf.Name, value)
			}
		}
	}
}

func parseRib(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	// RIB
	measurement := encodingPath
	for _, subfield := range field.Fields {
		// For Every table fill the keys which are vrfName, address and masklen
		switch subfield.Name {
		case "vrfName", "address", "maskLen":
			tags[subfield.Name] = decodeTag(subfield)
		}
		if value := decode(subfield); value != nil {
			grouper.Add(measurement, tags, timestamp, subfield.Name, value)
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
					grouper.Add(measurement, tags, timestamp, name, value)
				}
			}
		}
	}
}
