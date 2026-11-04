//go:build linux

package journald

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leodido/go-syslog/v4/common"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var priorityNames = map[string]bool{
	"emerg":   true,
	"alert":   true,
	"crit":    true,
	"err":     true,
	"warning": true,
	"notice":  true,
	"info":    true,
	"debug":   true,
}

// Journal fields always added as tags
var defaultTags = map[string]string{
	"_HOSTNAME":          "hostname",
	"_SYSTEMD_UNIT":      "unit",
	"_SYSTEMD_USER_UNIT": "user_unit",
	"SYSLOG_IDENTIFIER":  "identifier",
}

// checkPriority validates the priority setting in the forms accepted by
// journalctl, i.e. a single name or number or a range of those.
func checkPriority(p string) error {
	if p == "" {
		return nil
	}
	parts := strings.Split(p, "..")
	if len(parts) > 2 {
		return fmt.Errorf("invalid priority %q", p)
	}
	for _, part := range parts {
		if priorityNames[part] {
			continue
		}
		if v, err := strconv.Atoi(part); err == nil && v >= 0 && v <= 7 {
			continue
		}
		return fmt.Errorf("invalid priority %q", p)
	}
	return nil
}

// keyName converts a journal field name to a tag or field key
func keyName(name string) string {
	return strings.ToLower(strings.TrimLeft(name, "_"))
}

// decodeValue converts a journal field value of the JSON output to a string.
// journalctl encodes values as strings, binary values as arrays of bytes and
// fields occurring multiple times as an array of those, in which case the
// first value is used. Values that are too large to be shown are encoded as
// null and are skipped.
func decodeValue(raw json.RawMessage) (string, bool, error) {
	if len(raw) == 0 {
		return "", false, nil
	}
	switch raw[0] {
	case '"':
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", false, err
		}
		return s, true, nil
	case 'n':
		return "", false, nil
	case '[':
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return "", false, err
		}
		if len(values) == 0 {
			return "", true, nil
		}
		// Arrays of strings or arrays represent multiple values
		if c := values[0][0]; c == '"' || c == '[' || c == 'n' {
			return decodeValue(values[0])
		}
		// Otherwise the array contains the bytes of a binary value
		buf := make([]byte, 0, len(values))
		for _, v := range values {
			b, err := strconv.ParseUint(string(v), 10, 8)
			if err != nil {
				return "", false, fmt.Errorf("invalid byte %q in binary value", string(v))
			}
			buf = append(buf, byte(b))
		}
		return string(buf), true, nil
	}
	return "", false, fmt.Errorf("unexpected value %q", string(raw))
}

// parse converts a journal entry in JSON format to a metric and returns the
// cursor of the entry
func (j *Journald) parse(line []byte) (telegraf.Metric, string, error) {
	var entry map[string]json.RawMessage
	if err := json.Unmarshal(line, &entry); err != nil {
		return nil, "", fmt.Errorf("parsing journal entry failed: %w", err)
	}

	get := func(name string) (string, bool) {
		raw, found := entry[name]
		if !found {
			return "", false
		}
		v, ok, err := decodeValue(raw)
		if err != nil {
			j.Log.Debugf("Decoding journal field %q failed: %v", name, err)
			return "", false
		}
		return v, ok
	}

	cursor, _ := get("__CURSOR")

	ts := time.Now()
	if v, ok := get("__REALTIME_TIMESTAMP"); ok {
		usec, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, cursor, fmt.Errorf("invalid timestamp %q in journal entry", v)
		}
		ts = time.UnixMicro(usec)
	}

	tags := make(map[string]string, 6+len(j.tagKeys))
	fields := make(map[string]any, 4)

	message, _ := get("MESSAGE")
	fields["message"] = message

	if v, ok := get("PRIORITY"); ok {
		if code, err := strconv.ParseUint(v, 10, 8); err == nil {
			fields["severity_code"] = int64(code)
			if name, found := common.SeverityLevelsShort[uint8(code)]; found {
				tags["severity"] = name
			}
		}
	}
	if v, ok := get("SYSLOG_FACILITY"); ok {
		if code, err := strconv.ParseUint(v, 10, 8); err == nil {
			fields["facility_code"] = int64(code)
			if name, found := common.FacilityKeywords[uint8(code)]; found {
				tags["facility"] = name
			}
		}
	}
	if v, ok := get("_PID"); ok {
		if pid, err := strconv.ParseInt(v, 10, 64); err == nil {
			fields["pid"] = pid
		}
	}
	for name, key := range defaultTags {
		if v, ok := get(name); ok && v != "" {
			tags[key] = v
		}
	}

	// Add the user-requested tags and fields
	for name, key := range j.tagKeys {
		if v, ok := get(name); ok && v != "" {
			tags[key] = v
		}
	}
	if j.fieldKeys != nil {
		for name := range entry {
			if _, isTag := j.tagKeys[name]; isTag || !j.fieldKeys.Match(name) {
				continue
			}
			key := keyName(name)
			if _, exists := fields[key]; exists {
				continue
			}
			if _, exists := tags[key]; exists {
				continue
			}
			if v, ok := get(name); ok {
				fields[key] = v
			}
		}
	}

	return metric.New("journald", tags, fields, ts), cursor, nil
}
