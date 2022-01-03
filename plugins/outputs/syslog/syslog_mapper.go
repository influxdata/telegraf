package syslog

import (
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/go-syslog/v3/rfc5424"

	"github.com/influxdata/telegraf"
)

type SyslogMapper struct {
	DefaultSdid         string
	DefaultSeverityCode uint8
	DefaultFacilityCode uint8
	DefaultAppname      string
	Sdids               []string
	Separator           string
	reservedKeys        map[string]bool
}

// MapMetricToSyslogMessage maps metrics tags/fields to syslog messages
func (sm *SyslogMapper) MapMetricToSyslogMessage(metric telegraf.Metric) (*rfc5424.SyslogMessage, error) {
	msg := &rfc5424.SyslogMessage{}

	sm.mapPriority(metric, msg)
	sm.mapStructuredData(metric, msg)
	sm.mapAppname(metric, msg)
	mapHostname(metric, msg)
	mapTimestamp(metric, msg)
	mapMsgID(metric, msg)
	mapVersion(metric, msg)
	mapProcID(metric, msg)
	mapMsg(metric, msg)

	if !msg.Valid() {
		return nil, errors.New("metric could not produce valid syslog message")
	}
	return msg, nil
}

func (sm *SyslogMapper) mapStructuredData(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	for _, tag := range metric.TagList() {
		sm.mapStructuredDataItem(tag.Key, tag.Value, msg)
	}
	for _, field := range metric.FieldList() {
		sm.mapStructuredDataItem(field.Key, formatValue(field.Value), msg)
	}
}

func (sm *SyslogMapper) mapStructuredDataItem(key string, value string, msg *rfc5424.SyslogMessage) {
	if sm.reservedKeys[key] {
		return
	}
	isExplicitSdid := false
	for _, sdid := range sm.Sdids {
		k := strings.TrimLeft(key, sdid+sm.Separator)
		if len(key) > len(k) {
			isExplicitSdid = true
			msg.SetParameter(sdid, k, value)
			break
		}
	}
	if !isExplicitSdid && len(sm.DefaultSdid) > 0 {
		k := strings.TrimPrefix(key, sm.DefaultSdid+sm.Separator)
		msg.SetParameter(sm.DefaultSdid, k, value)
	}
}

func (sm *SyslogMapper) mapAppname(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	if value, ok := metric.GetTag("appname"); ok {
		msg.SetAppname(formatValue(value))
	} else {
		//Use default appname
		msg.SetAppname(sm.DefaultAppname)
	}
}

func mapMsgID(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	if value, ok := metric.GetField("msgid"); ok {
		msg.SetMsgID(formatValue(value))
	} else {
		// We default to metric name
		msg.SetMsgID(metric.Name())
	}
}

func mapVersion(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	if value, ok := metric.GetField("version"); ok {
		if v, ok := value.(uint64); ok {
			msg.SetVersion(uint16(v))
			return
		}
	}
	msg.SetVersion(1)
}

func mapMsg(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	if value, ok := metric.GetField("msg"); ok {
		msg.SetMessage(formatValue(value))
	}
}

func mapProcID(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	if value, ok := metric.GetField("procid"); ok {
		msg.SetProcID(formatValue(value))
	}
}

func (sm *SyslogMapper) mapPriority(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	severityCode := sm.DefaultSeverityCode
	facilityCode := sm.DefaultFacilityCode

	if value, ok := getFieldCode(metric, "severity_code"); ok {
		severityCode = *value
	}

	if value, ok := getFieldCode(metric, "facility_code"); ok {
		facilityCode = *value
	}

	priority := (8 * facilityCode) + severityCode
	msg.SetPriority(priority)
}

func mapHostname(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	// Try with hostname, then with source, then with host tags, then take OS Hostname
	if value, ok := metric.GetTag("hostname"); ok {
		msg.SetHostname(formatValue(value))
	} else if value, ok := metric.GetTag("source"); ok {
		msg.SetHostname(formatValue(value))
	} else if value, ok := metric.GetTag("host"); ok {
		msg.SetHostname(formatValue(value))
	} else if value, err := os.Hostname(); err == nil {
		msg.SetHostname(value)
	}
}

func mapTimestamp(metric telegraf.Metric, msg *rfc5424.SyslogMessage) {
	timestamp := metric.Time()
	//nolint: revive // Need switch with only one case to handle `.(type)`
	if value, ok := metric.GetField("timestamp"); ok {
		if v, ok := value.(int64); ok {
			timestamp = time.Unix(0, v).UTC()
		}
	}
	msg.SetTimestamp(timestamp.Format(time.RFC3339))
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "1"
		}
		return "0"
	case uint64:
		return strconv.FormatUint(v, 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		if math.IsNaN(v) {
			return ""
		}

		if math.IsInf(v, 0) {
			return ""
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	}

	return ""
}

func getFieldCode(metric telegraf.Metric, fieldKey string) (*uint8, bool) {
	if value, ok := metric.GetField(fieldKey); ok {
		if v, err := strconv.ParseUint(formatValue(value), 10, 8); err == nil {
			r := uint8(v)
			return &r, true
		}
	}
	return nil, false
}

func newSyslogMapper() *SyslogMapper {
	return &SyslogMapper{
		reservedKeys: map[string]bool{
			"version": true, "severity_code": true, "facility_code": true,
			"procid": true, "msgid": true, "msg": true, "timestamp": true, "sdid": true,
			"hostname": true, "source": true, "host": true, "severity": true,
			"facility": true, "appname": true},
	}
}
