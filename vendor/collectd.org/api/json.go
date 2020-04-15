package api

import (
	"encoding/json"
	"fmt"

	"collectd.org/cdtime"
)

// jsonValueList represents the format used by collectd's JSON export.
type jsonValueList struct {
	Values         []json.Number `json:"values"`
	DSTypes        []string      `json:"dstypes"`
	DSNames        []string      `json:"dsnames,omitempty"`
	Time           cdtime.Time   `json:"time"`
	Interval       cdtime.Time   `json:"interval"`
	Host           string        `json:"host"`
	Plugin         string        `json:"plugin"`
	PluginInstance string        `json:"plugin_instance,omitempty"`
	Type           string        `json:"type"`
	TypeInstance   string        `json:"type_instance,omitempty"`
}

// MarshalJSON implements the "encoding/json".Marshaler interface for
// ValueList.
func (vl *ValueList) MarshalJSON() ([]byte, error) {
	jvl := jsonValueList{
		Values:         make([]json.Number, len(vl.Values)),
		DSTypes:        make([]string, len(vl.Values)),
		DSNames:        make([]string, len(vl.Values)),
		Time:           cdtime.New(vl.Time),
		Interval:       cdtime.NewDuration(vl.Interval),
		Host:           vl.Host,
		Plugin:         vl.Plugin,
		PluginInstance: vl.PluginInstance,
		Type:           vl.Type,
		TypeInstance:   vl.TypeInstance,
	}

	for i, v := range vl.Values {
		switch v := v.(type) {
		case Gauge:
			jvl.Values[i] = json.Number(fmt.Sprintf("%.15g", v))
		case Derive:
			jvl.Values[i] = json.Number(fmt.Sprintf("%d", v))
		case Counter:
			jvl.Values[i] = json.Number(fmt.Sprintf("%d", v))
		default:
			return nil, fmt.Errorf("unexpected data source type: %T", v)
		}
		jvl.DSTypes[i] = v.Type()
		jvl.DSNames[i] = vl.DSName(i)
	}

	return json.Marshal(jvl)
}

// UnmarshalJSON implements the "encoding/json".Unmarshaler interface for
// ValueList.
//
// Please note that this function is currently not compatible with write_http's
// "StoreRates" setting: if enabled, write_http converts derives and counters
// to a rate (a floating point number), but still puts "derive" or "counter" in
// the "dstypes" array. UnmarshalJSON will try to parse such values as
// integers, which will fail in many cases.
func (vl *ValueList) UnmarshalJSON(data []byte) error {
	var jvl jsonValueList

	if err := json.Unmarshal(data, &jvl); err != nil {
		return err
	}

	vl.Host = jvl.Host
	vl.Plugin = jvl.Plugin
	vl.PluginInstance = jvl.PluginInstance
	vl.Type = jvl.Type
	vl.TypeInstance = jvl.TypeInstance

	vl.Time = jvl.Time.Time()
	vl.Interval = jvl.Interval.Duration()
	vl.Values = make([]Value, len(jvl.Values))

	if len(jvl.Values) != len(jvl.DSTypes) {
		return fmt.Errorf("invalid data: %d value(s), %d data source type(s)",
			len(jvl.Values), len(jvl.DSTypes))
	}

	for i, n := range jvl.Values {
		switch jvl.DSTypes[i] {
		case "gauge":
			v, err := n.Float64()
			if err != nil {
				return err
			}
			vl.Values[i] = Gauge(v)
		case "derive":
			v, err := n.Int64()
			if err != nil {
				return err
			}
			vl.Values[i] = Derive(v)
		case "counter":
			v, err := n.Int64()
			if err != nil {
				return err
			}
			vl.Values[i] = Counter(v)
		default:
			return fmt.Errorf("unexpected data source type: %q", jvl.DSTypes[i])
		}
	}

	if len(jvl.DSNames) >= len(vl.Values) {
		vl.DSNames = make([]string, len(vl.Values))
		copy(vl.DSNames, jvl.DSNames)
	}

	return nil
}
