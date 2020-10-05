package devo

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
)

type DevoMapper struct {
	DefaultHostname     string
	DefaultSeverityCode uint8
	DefaultFacilityCode uint8
	DefaultTag          string
}

func (ds *DevoMapper) devoMapper(metric telegraf.Metric, msg []byte) ([]byte, error) {
	err := error(nil)

	devoTag := ds.DefaultTag
	severityCode := ds.DefaultSeverityCode
	facilityCode := ds.DefaultFacilityCode
	hostname := ds.DefaultHostname

	if value, ok := metric.GetTag("devo_tag"); ok {
		devoTag = formatValue(value)
	}

	if value, ok := getFieldCode(metric, "severity_code"); ok {
		severityCode = *value
	}

	if value, ok := getFieldCode(metric, "facility_code"); ok {
		facilityCode = *value
	}

	priority := strconv.Itoa((int((8 * facilityCode) + severityCode)))

	if value, ok := metric.GetTag("hostname"); ok {
		hostname = formatValue(value)
	} else if value, ok := metric.GetTag("source"); ok {
		hostname = formatValue(value)
	} else if value, ok := metric.GetTag("host"); ok {
		hostname = formatValue(value)
	} else if value, err := os.Hostname(); err == nil {
		hostname = value
	}

	timestamp := metric.Time()
	if value, ok := metric.GetField("timestamp"); ok {
		switch v := value.(type) {
		case int64:
			timestamp = time.Unix(0, v).UTC()
		}
	}
	sendTime := timestamp.Format(time.RFC3339)

	devomsg := fmt.Sprintf("<%s>%s %s %s: %s", priority, sendTime, hostname, devoTag, msg)

	return []byte(devomsg), err
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
